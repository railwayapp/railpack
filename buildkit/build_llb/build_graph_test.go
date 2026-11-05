package build_llb

import (
	"testing"

	"github.com/moby/buildkit/client/llb"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphNoCache(t *testing.T) {
	p := &plan.BuildPlan{
		Steps: []plan.Step{
			{
				Name:   "test",
				Caches: []string{"test-cache"},
			},
		},
		Caches: map[string]*plan.Cache{
			"test-cache": {
				Directory: "/root/.cache",
			},
		},
	}

	localState := llb.Local("context")
	cacheStore := NewBuildKitCacheStore("")
	platform := specs.Platform{OS: "linux", Architecture: "amd64"}

	t.Run("with NoCache=false", func(t *testing.T) {
		g, err := NewBuildGraph(p, &localState, cacheStore, "", &platform, "", false)
		require.NoError(t, err)
		require.False(t, g.NoCache)
	})

	t.Run("with NoCache=true", func(t *testing.T) {
		g, err := NewBuildGraph(p, &localState, cacheStore, "", &platform, "", true)
		require.NoError(t, err)
		require.True(t, g.NoCache)
		// NoCache should NOT affect if caches are enabled or not, it just affects the layer cache
		opts, err := g.getCacheMountOptions([]string{"test-cache"})
		require.NoError(t, err)
		require.Len(t, opts, 1)
	})
}

func TestBuildGraphConnectsStepInputs(t *testing.T) {
	p := &plan.BuildPlan{
		Steps: []plan.Step{
			{Name: "A"},
			{Name: "B", Inputs: []plan.Layer{plan.NewImageLayer("alpine:latest"), plan.NewStepLayer("A")}},
			{Name: "C", Inputs: []plan.Layer{plan.NewStepLayer("A"), plan.NewStepLayer("B")}},
		},
	}
	localState := llb.Local("context")
	platform := specs.Platform{OS: "linux", Architecture: "amd64"}
	g, err := NewBuildGraph(p, &localState, NewBuildKitCacheStore(""), "", &platform, "", false)
	require.NoError(t, err)

	a, exists := g.graph.nodes["A"]
	require.True(t, exists)
	b, exists := g.graph.nodes["B"]
	require.True(t, exists)
	c, exists := g.graph.nodes["C"]
	require.True(t, exists)

	// Removing A -> C must update both directions while preserving A -> B -> C.
	require.Empty(t, a.parents)
	require.Equal(t, []*StepNode{a}, b.parents)
	require.Equal(t, []*StepNode{b}, c.parents)
	require.Equal(t, []*StepNode{b}, a.children)
	require.Equal(t, []*StepNode{c}, b.children)
	require.Empty(t, c.children)
}

func newTestStepNode(name string) *StepNode {
	return &StepNode{Step: &plan.Step{Name: name}}
}

func TestGraphBasicOperations(t *testing.T) {
	g := NewGraph()

	nodeA := newTestStepNode("A")
	nodeB := newTestStepNode("B")

	g.nodes[nodeA.Step.Name] = nodeA
	g.nodes[nodeB.Step.Name] = nodeB

	if len(g.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g.nodes))
	}

	if node, exists := g.nodes["A"]; !exists || node != nodeA {
		t.Error("Failed to retrieve node A")
	}

	if _, exists := g.nodes["C"]; exists {
		t.Error("Retrieved non-existent node")
	}
}

func TestGraphProcessingOrder(t *testing.T) {
	tests := []struct {
		name  string
		nodes []string
		edges [][2]string
	}{
		{name: "empty"},
		{name: "single node", nodes: []string{"A"}},
		{name: "chain", nodes: []string{"A", "B", "C"}, edges: [][2]string{{"A", "B"}, {"B", "C"}}},
		{
			name:  "diamond",
			nodes: []string{"A", "B", "C", "D"},
			edges: [][2]string{{"A", "B"}, {"A", "C"}, {"B", "D"}, {"C", "D"}},
		},
		{
			name:  "shared descendant",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{{"A", "B"}, {"A", "C"}, {"B", "C"}},
		},
		{
			name:  "disconnected components and isolated node",
			nodes: []string{"A", "B", "C", "D", "E"},
			edges: [][2]string{{"A", "B"}, {"C", "D"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := make(map[string]*StepNode, len(tt.nodes))
			for _, name := range tt.nodes {
				nodes[name] = newTestStepNode(name)
			}
			// Outgoing edges are sufficient to describe the graph for topological sorting.
			for _, edge := range tt.edges {
				nodes[edge[0]].children = append(nodes[edge[0]].children, nodes[edge[1]])
			}

			order, err := computeProcessingOrder(nodes)
			require.NoError(t, err)
			require.Len(t, order, len(nodes))

			positions := make(map[string]int, len(order))
			for i, node := range order {
				require.NotContains(t, positions, node.Step.Name)
				require.Same(t, nodes[node.Step.Name], node)
				positions[node.Step.Name] = i
			}
			// Independent nodes may have different valid orders because nodes is a map.
			for _, edge := range tt.edges {
				require.Less(t, positions[edge[0]], positions[edge[1]], "dependency %s must precede %s", edge[0], edge[1])
			}
		})
	}
}

func TestGraphCycleDetection(t *testing.T) {
	tests := []struct {
		name  string
		nodes []string
		edges [][2]string
	}{
		{name: "self loop", nodes: []string{"A"}, edges: [][2]string{{"A", "A"}}},
		{name: "two node cycle", nodes: []string{"A", "B"}, edges: [][2]string{{"A", "B"}, {"B", "A"}}},
		{
			name:  "three node cycle without leaves",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{{"A", "B"}, {"B", "C"}, {"C", "A"}},
		},
		{
			name:  "cycle disconnected from acyclic component",
			nodes: []string{"A", "B", "C", "D", "E"},
			edges: [][2]string{{"A", "B"}, {"C", "D"}, {"D", "E"}, {"E", "C"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := make(map[string]*StepNode, len(tt.nodes))
			for _, name := range tt.nodes {
				nodes[name] = newTestStepNode(name)
			}
			// A cycle has no leaf to use as a traversal entry point.
			for _, edge := range tt.edges {
				nodes[edge[0]].children = append(nodes[edge[0]].children, nodes[edge[1]])
			}

			order, err := computeProcessingOrder(nodes)
			require.Error(t, err)
			require.Nil(t, order)
		})
	}
}

func TestTransitiveDependencies(t *testing.T) {
	// Create a graph with redundant edges:
	//   A
	//  /| \
	// B |  C
	//  \| / \
	//   D    E
	nodes := map[string]*StepNode{
		"A": newTestStepNode("A"),
		"B": newTestStepNode("B"),
		"C": newTestStepNode("C"),
		"D": newTestStepNode("D"),
		"E": newTestStepNode("E"),
	}

	nodes["B"].parents = []*StepNode{nodes["A"]}
	nodes["C"].parents = []*StepNode{nodes["A"]}
	nodes["D"].parents = []*StepNode{nodes["A"], nodes["B"], nodes["C"]} // A is redundant
	nodes["E"].parents = []*StepNode{nodes["C"]}

	nodes["A"].children = []*StepNode{nodes["B"], nodes["C"], nodes["D"]}
	nodes["B"].children = []*StepNode{nodes["D"]}
	nodes["C"].children = []*StepNode{nodes["D"], nodes["E"]}

	// Remove redundant edges
	computeTransitiveDependencies(nodes)

	// Verify D's parents (should only have B and C as parents)
	dParents := nodes["D"].parents
	if len(dParents) != 2 {
		t.Errorf("Expected 2 parents for D after transitive reduction, got %d", len(dParents))
	}

	// Verify A is not a direct parent of D
	for _, parent := range dParents {
		if parent.Step.Name == "A" {
			t.Error("Node A should not be a direct parent of D after transitive reduction")
		}
	}
}

func TestTransitiveDependenciesRemovesRedundantEdges(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []string
		edges     [][2]string
		wantEdges [][2]string
	}{
		{
			/*
				┌───────►B──────►C
				│        │       │
				A        │       │
				|        ▼       │
				└──────► D ◄─────┘
			*/
			name:      "chain with shortcuts to D",
			nodes:     []string{"A", "B", "C", "D"},
			edges:     [][2]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"B", "D"}, {"A", "D"}},
			wantEdges: [][2]string{{"A", "B"}, {"B", "C"}, {"C", "D"}},
		},
		{
			/*
				A ──► B ──► C
				│     │╲    │
				│     │ ╲   │
				│     │  ╲  │
				│     │   ╲ │
				│     ▼    ↘▼
				│     D ──► E
				│           ▲
				└───────────┘
			*/
			name:      "diamond with shortcuts to E",
			nodes:     []string{"A", "B", "C", "D", "E"},
			edges:     [][2]string{{"A", "B"}, {"B", "C"}, {"B", "D"}, {"B", "E"}, {"C", "E"}, {"D", "E"}, {"A", "E"}},
			wantEdges: [][2]string{{"A", "B"}, {"B", "C"}, {"B", "D"}, {"C", "E"}, {"D", "E"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := make(map[string]*StepNode, len(tt.nodes))
			for _, name := range tt.nodes {
				nodes[name] = newTestStepNode(name)
			}
			for _, edge := range tt.edges {
				parent, child := nodes[edge[0]], nodes[edge[1]]
				child.parents = append(child.parents, parent)
				parent.children = append(parent.children, child)
			}

			computeTransitiveDependencies(nodes)

			require.Len(t, nodes, len(tt.nodes))
			// Check both directions so removed dependencies cannot leave stale child edges.
			var parentEdges, childEdges [][2]string
			for _, name := range tt.nodes {
				node, exists := nodes[name]
				require.True(t, exists, "node %s must be preserved", name)
				for _, parent := range node.parents {
					parentEdges = append(parentEdges, [2]string{parent.Step.Name, name})
				}
				for _, child := range node.children {
					childEdges = append(childEdges, [2]string{name, child.Step.Name})
				}
			}
			require.ElementsMatch(t, tt.wantEdges, parentEdges, "parent edges")
			require.ElementsMatch(t, tt.wantEdges, childEdges, "child edges")
		})
	}
}

func TestTransitiveDependenciesWithDistinctPointersToIdenticalNodes(t *testing.T) {
	nodeA := newTestStepNode("A")
	nodeB := newTestStepNode("B")
	nodeC := newTestStepNode("C")

	nodeA.children = []*StepNode{nodeB, nodeC}
	nodeB.children = []*StepNode{nodeC}

	// Copied nodes have distinct pointer identities even when they refer to the same step.
	nodeACopy := *nodeA
	directParent := nodeA
	indirectParent := &nodeACopy
	if directParent == indirectParent {
		t.Fatal("Expected distinct node pointers to compare unequal")
	}

	// A -> C is redundant because B depends on an identical copy of A.
	nodeB.parents = []*StepNode{indirectParent}
	nodeC.parents = []*StepNode{directParent, nodeB}

	nodes := map[string]*StepNode{
		nodeA.Step.Name: nodeA,
		nodeB.Step.Name: nodeB,
		nodeC.Step.Name: nodeC,
	}

	computeTransitiveDependencies(nodes)

	cParents := nodeC.parents
	if len(cParents) != 1 || cParents[0].Step.Name != "B" {
		t.Errorf("Expected only B as C's parent after transitive reduction, got %v", cParents)
	}

	aChildren := nodeA.children
	if len(aChildren) != 1 || aChildren[0].Step.Name != "B" {
		t.Errorf("Expected only B as A's child after transitive reduction, got %v", aChildren)
	}
}
