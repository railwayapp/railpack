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
	g := NewGraph()

	// Create a simple graph:
	//   A
	//  / \
	// B   C
	//  \ /
	//   D
	nodeA := newTestStepNode("A")
	nodeB := newTestStepNode("B")
	nodeC := newTestStepNode("C")
	nodeD := newTestStepNode("D")

	g.nodes[nodeA.Step.Name] = nodeA
	g.nodes[nodeB.Step.Name] = nodeB
	g.nodes[nodeC.Step.Name] = nodeC
	g.nodes[nodeD.Step.Name] = nodeD

	nodeB.parents = []*StepNode{nodeA}
	nodeC.parents = []*StepNode{nodeA}
	nodeD.parents = []*StepNode{nodeB, nodeC}

	nodeA.children = []*StepNode{nodeB, nodeC}
	nodeB.children = []*StepNode{nodeD}
	nodeC.children = []*StepNode{nodeD}

	order, err := g.ComputeProcessingOrder()
	if err != nil {
		t.Fatalf("Failed to compute processing order: %v", err)
	}

	names := make([]string, len(order))
	for i, node := range order {
		names[i] = node.Step.Name
	}
	t.Logf("Order: %v", names)

	// Verify order (should be A before B and C, and B and C before D)
	if len(order) != 4 {
		t.Fatalf("Expected 4 nodes in order, got %d", len(order))
	}

	// A should be first
	if order[0].Step.Name != "A" {
		t.Errorf("Expected A to be first, got %s", order[0].Step.Name)
	}

	// D should be last
	if order[3].Step.Name != "D" {
		t.Errorf("Expected D to be last, got %s", order[3].Step.Name)
	}
}

func TestGraphCycleDetection(t *testing.T) {
	g := NewGraph()

	// Create a cyclic graph:
	// A -> B -> C -> A
	nodeA := newTestStepNode("A")
	nodeB := newTestStepNode("B")
	nodeC := newTestStepNode("C")

	g.nodes[nodeA.Step.Name] = nodeA
	g.nodes[nodeB.Step.Name] = nodeB
	g.nodes[nodeC.Step.Name] = nodeC

	nodeB.parents = []*StepNode{nodeA}
	nodeC.parents = []*StepNode{nodeB}
	nodeA.parents = []*StepNode{nodeC}

	nodeA.children = []*StepNode{nodeB}
	nodeB.children = []*StepNode{nodeC}
	nodeC.children = []*StepNode{nodeA}

	// Test cycle detection
	_, err := g.ComputeProcessingOrder()
	if err == nil {
		t.Error("Expected cycle detection error, got nil")
	}
}

func TestTransitiveDependencies(t *testing.T) {
	g := NewGraph()

	// Create a graph with redundant edges:
	//   A
	//  / \
	// B   C
	//  \ / \
	//   D   E
	nodeA := newTestStepNode("A")
	nodeB := newTestStepNode("B")
	nodeC := newTestStepNode("C")
	nodeD := newTestStepNode("D")
	nodeE := newTestStepNode("E")

	g.nodes[nodeA.Step.Name] = nodeA
	g.nodes[nodeB.Step.Name] = nodeB
	g.nodes[nodeC.Step.Name] = nodeC
	g.nodes[nodeD.Step.Name] = nodeD
	g.nodes[nodeE.Step.Name] = nodeE

	nodeB.parents = []*StepNode{nodeA}
	nodeC.parents = []*StepNode{nodeA}
	nodeD.parents = []*StepNode{nodeA, nodeB, nodeC} // A is redundant
	nodeE.parents = []*StepNode{nodeC}

	nodeA.children = []*StepNode{nodeB, nodeC, nodeD}
	nodeB.children = []*StepNode{nodeD}
	nodeC.children = []*StepNode{nodeD, nodeE}

	// Remove redundant edges
	g.ComputeTransitiveDependencies()

	// Verify D's parents (should only have B and C as parents)
	dParents := nodeD.parents
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

func TestTransitiveDependenciesWithDistinctPointersToIdenticalNodes(t *testing.T) {
	g := NewGraph()

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

	g.nodes[nodeA.Step.Name] = nodeA
	g.nodes[nodeB.Step.Name] = nodeB
	g.nodes[nodeC.Step.Name] = nodeC

	g.ComputeTransitiveDependencies()

	cParents := nodeC.parents
	if len(cParents) != 1 || cParents[0].Step.Name != "B" {
		t.Errorf("Expected only B as C's parent after transitive reduction, got %v", cParents)
	}

	aChildren := nodeA.children
	if len(aChildren) != 1 || aChildren[0].Step.Name != "B" {
		t.Errorf("Expected only B as A's child after transitive reduction, got %v", aChildren)
	}
}
