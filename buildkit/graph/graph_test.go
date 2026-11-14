package graph

import (
	"testing"
)

// TestNode is a simple implementation of Node for testing
type TestNode struct {
	name     string
	parents  []Node
	children []Node
}

func NewTestNode(name string) *TestNode {
	return &TestNode{
		name:     name,
		parents:  make([]Node, 0),
		children: make([]Node, 0),
	}
}

func (n *TestNode) GetName() string      { return n.name }
func (n *TestNode) GetParents() []Node   { return n.parents }
func (n *TestNode) GetChildren() []Node  { return n.children }
func (n *TestNode) SetParents(p []Node)  { n.parents = p }
func (n *TestNode) SetChildren(c []Node) { n.children = c }

func TestGraphBasicOperations(t *testing.T) {
	g := NewGraph()

	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")

	g.AddNode(nodeA)
	g.AddNode(nodeB)

	if len(g.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g.GetNodes()))
	}

	if node, exists := g.GetNode("A"); !exists || node != nodeA {
		t.Error("Failed to retrieve node A")
	}

	if _, exists := g.GetNode("C"); exists {
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
	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")
	nodeD := NewTestNode("D")

	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)
	g.AddNode(nodeD)

	nodeB.SetParents([]Node{nodeA})
	nodeC.SetParents([]Node{nodeA})
	nodeD.SetParents([]Node{nodeB, nodeC})

	nodeA.SetChildren([]Node{nodeB, nodeC})
	nodeB.SetChildren([]Node{nodeD})
	nodeC.SetChildren([]Node{nodeD})

	order, err := g.ComputeProcessingOrder()
	if err != nil {
		t.Fatalf("Failed to compute processing order: %v", err)
	}

	names := make([]string, len(order))
	for i, node := range order {
		names[i] = node.GetName()
	}
	t.Logf("Order: %v", names)

	// Verify order (should be A before B and C, and B and C before D)
	if len(order) != 4 {
		t.Fatalf("Expected 4 nodes in order, got %d", len(order))
	}

	// A should be first
	if order[0].GetName() != "A" {
		t.Errorf("Expected A to be first, got %s", order[0].GetName())
	}

	// D should be last
	if order[3].GetName() != "D" {
		t.Errorf("Expected D to be last, got %s", order[3].GetName())
	}
}

func TestGraphCycleDetection(t *testing.T) {
	g := NewGraph()

	// Create a cyclic graph:
	// A -> B -> C -> A
	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")

	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)

	nodeB.SetParents([]Node{nodeA})
	nodeC.SetParents([]Node{nodeB})
	nodeA.SetParents([]Node{nodeC})

	nodeA.SetChildren([]Node{nodeB})
	nodeB.SetChildren([]Node{nodeC})
	nodeC.SetChildren([]Node{nodeA})

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
	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")
	nodeD := NewTestNode("D")
	nodeE := NewTestNode("E")

	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)
	g.AddNode(nodeD)
	g.AddNode(nodeE)

	nodeB.SetParents([]Node{nodeA})
	nodeC.SetParents([]Node{nodeA})
	nodeD.SetParents([]Node{nodeA, nodeB, nodeC}) // A is redundant
	nodeE.SetParents([]Node{nodeC})

	nodeA.SetChildren([]Node{nodeB, nodeC, nodeD})
	nodeB.SetChildren([]Node{nodeD})
	nodeC.SetChildren([]Node{nodeD, nodeE})

	// Remove redundant edges
	if err := g.ComputeTransitiveDependencies(); err != nil {
		t.Fatalf("ComputeTransitiveDependencies failed: %v", err)
	}

	// Verify D's parents (should only have B and C as parents)
	dParents := nodeD.GetParents()
	if len(dParents) != 2 {
		t.Errorf("Expected 2 parents for D after transitive reduction, got %d", len(dParents))
	}

	// Verify A is not a direct parent of D
	for _, parent := range dParents {
		if parent.GetName() == "A" {
			t.Error("Node A should not be a direct parent of D after transitive reduction")
		}
	}
}

func TestTransitiveDependencies_CycleDetection(t *testing.T) {
	g := NewGraph()

	// Cyclic graph:
	// A.parents = [B, C]
	// B.parents = [A]
	// C.parents = [A]
	// Transitive reduction without prior cycle detection would treat B as reachable
	// via C -> A -> B and C as reachable via B -> A -> C, erasing A's parents and masking the cycle.
	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")

	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)

	nodeA.SetParents([]Node{nodeB, nodeC})
	nodeB.SetParents([]Node{nodeA})
	nodeC.SetParents([]Node{nodeA})

	nodeA.SetChildren([]Node{nodeB, nodeC})
	nodeB.SetChildren([]Node{nodeA})
	nodeC.SetChildren([]Node{nodeA})

	err := g.ComputeTransitiveDependencies()
	if err == nil {
		t.Fatal("Expected cycle detection error from ComputeTransitiveDependencies, got nil")
	}

	// Verify cycle is not erased: A still has its parents
	if len(nodeA.GetParents()) != 2 {
		t.Fatalf("Expected node A to retain 2 parents, got %d", len(nodeA.GetParents()))
	}

	// Verify ComputeProcessingOrder also still detects the cycle
	_, orderErr := g.ComputeProcessingOrder()
	if orderErr == nil {
		t.Fatal("Expected cycle detection error from ComputeProcessingOrder, got nil")
	}
}
