package graph

import (
	"slices"
	"testing"
)

func TestTransitiveDependenciesPreservesCycles(t *testing.T) {
	cases := []struct {
		name  string
		nodes []string
		edges [][2]string
	}{
		{
			name:  "self loop",
			nodes: []string{"A"},
			edges: [][2]string{{"A", "A"}},
		},
		{
			name:  "two node cycle",
			nodes: []string{"A", "B"},
			edges: [][2]string{{"A", "B"}, {"B", "A"}},
		},
		{
			name:  "simple cycle",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{{"A", "B"}, {"B", "C"}, {"C", "A"}},
		},
		{
			name:  "shared cycles",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{{"A", "B"}, {"B", "A"}, {"A", "C"}, {"C", "A"}},
		},
		{
			name:  "disconnected cycle",
			nodes: []string{"A", "B", "C", "D", "E"},
			edges: [][2]string{{"A", "B"}, {"C", "D"}, {"D", "E"}, {"E", "C"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGraph()
			nodes := make(map[string]*TestNode)
			for _, name := range tc.nodes {
				nodes[name] = NewTestNode(name)
				g.AddNode(nodes[name])
			}
			for _, edge := range tc.edges {
				parent, child := nodes[edge[0]], nodes[edge[1]]
				child.SetParents(append(child.GetParents(), parent))
				parent.SetChildren(append(parent.GetChildren(), child))
			}

			parents := make(map[string][]Node)
			children := make(map[string][]Node)
			for name, node := range nodes {
				parents[name] = slices.Clone(node.GetParents())
				children[name] = slices.Clone(node.GetChildren())
			}
			if _, err := g.ComputeProcessingOrder(); err == nil {
				t.Fatal("Expected a cycle in the original graph")
			}

			if err := g.ComputeTransitiveDependencies(); err == nil {
				t.Fatal("Expected transitive reduction to reject the cyclic graph")
			}

			if _, err := g.ComputeProcessingOrder(); err == nil {
				t.Error("Transitive reduction hid the cycle from processing-order validation")
			}
			for name, node := range nodes {
				if !slices.Equal(node.GetParents(), parents[name]) {
					t.Errorf("Parents of %s changed in a cyclic graph", name)
				}
				if !slices.Equal(node.GetChildren(), children[name]) {
					t.Errorf("Children of %s changed in a cyclic graph", name)
				}
			}
		})
	}
}
