package graph

import "testing"

func TestInterfaceEqualityBug_1(t *testing.T) {
	g := NewGraph()

	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")

	nodeA.SetChildren([]Node{nodeB, nodeC})
	nodeB.SetChildren([]Node{nodeC})

	// Interface equality compares pointer addresses, even when all struct fields match.
	nodeACopy := *nodeA
	var directParent Node = nodeA
	var indirectParent Node = &nodeACopy
	if directParent == indirectParent {
		t.Fatal("Expected distinct pointers to compare unequal as Node interfaces")
	}

	// A -> C is redundant because B depends on an identical copy of A.
	nodeB.SetParents([]Node{indirectParent})
	nodeC.SetParents([]Node{directParent, nodeB})

	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)

	g.ComputeTransitiveDependencies()

	cParents := nodeC.GetParents()
	if len(cParents) != 1 || cParents[0].GetName() != "B" {
		t.Errorf("Expected only B as C's parent after transitive reduction, got %v", cParents)
	}

	aChildren := nodeA.GetChildren()
	if len(aChildren) != 1 || aChildren[0].GetName() != "B" {
		t.Errorf("Expected only B as A's child after transitive reduction, got %v", aChildren)
	}
}

func TestInterfaceEqualityBug_2(t *testing.T) {
	g := NewGraph()
	var nodeA Node = NewTestNode("A")
	var anotherNodeA Node = NewTestNode("A")

	// Identical constructor arguments do not make separately allocated pointers equal.
	if nodeA == anotherNodeA {
		t.Fatal("Expected distinct pointers to compare unequal as Node interfaces")
	}

	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")
	nodeA.SetChildren([]Node{nodeB, nodeC})
	anotherNodeA.SetChildren([]Node{nodeB, nodeC})
	nodeB.SetParents([]Node{anotherNodeA})
	nodeB.SetChildren([]Node{nodeC})
	nodeC.SetParents([]Node{nodeA, nodeB})
	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)

	g.ComputeTransitiveDependencies()

	parents := nodeC.GetParents()
	if len(parents) != 1 || parents[0].GetName() != "B" {
		t.Errorf("Expected only B as C's parent, got %v", parents)
	}
	children := nodeA.GetChildren()
	if len(children) != 1 || children[0].GetName() != "B" {
		t.Errorf("Expected only B as A's child, got %v", children)
	}
}

func TestTransitiveDependenciesRemovesCopiedChildReference(t *testing.T) {
	g := NewGraph()
	nodeA := NewTestNode("A")
	nodeB := NewTestNode("B")
	nodeC := NewTestNode("C")
	nodeCCopy := *nodeC

	// The reverse edge represents C by another pointer to the same named node.
	nodeA.SetChildren([]Node{nodeB, &nodeCCopy})
	nodeB.SetParents([]Node{nodeA})
	nodeB.SetChildren([]Node{nodeC})
	nodeC.SetParents([]Node{nodeA, nodeB})
	g.AddNode(nodeA)
	g.AddNode(nodeB)
	g.AddNode(nodeC)

	g.ComputeTransitiveDependencies()

	parents := nodeC.GetParents()
	if len(parents) != 1 || parents[0].GetName() != "B" {
		t.Errorf("Expected only B as C's parent, got %v", parents)
	}
	children := nodeA.GetChildren()
	if len(children) != 1 || children[0].GetName() != "B" {
		t.Errorf("Expected redundant child C to be removed by name, got %v", children)
	}
}

func TestTransitiveDependenciesKeepsSameNameParentReferences(t *testing.T) {
	g := NewGraph()
	nodeA := NewTestNode("A")
	anotherNodeA := NewTestNode("A")
	nodeC := NewTestNode("C")
	nodeA.SetChildren([]Node{nodeC})
	anotherNodeA.SetChildren([]Node{nodeC})
	nodeC.SetParents([]Node{nodeA, anotherNodeA})
	g.AddNode(nodeA)
	g.AddNode(nodeC)

	g.ComputeTransitiveDependencies()

	// Same-name references must not make each other redundant and erase the dependency.
	parents := nodeC.GetParents()
	if len(parents) == 0 {
		t.Fatal("Expected C to retain its dependency on A")
	}
	for _, parent := range parents {
		if parent.GetName() != "A" {
			t.Errorf("Expected only A parent references, got %s", parent.GetName())
		}
	}
}
