package graph

import (
	"fmt"
)

// Node represents a node in a directed graph
type Node interface {
	GetName() string
	GetParents() []Node
	GetChildren() []Node
	SetParents([]Node)
	SetChildren([]Node)
}

// Graph represents a directed graph structure
type Graph struct {
	nodes map[string]Node
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]Node),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node Node) {
	g.nodes[node.GetName()] = node
}

// GetNode retrieves a node by name
func (g *Graph) GetNode(name string) (Node, bool) {
	node, exists := g.nodes[name]
	return node, exists
}

// GetNodes returns all nodes in the graph
func (g *Graph) GetNodes() map[string]Node {
	return g.nodes
}

// ComputeProcessingOrder returns nodes in topological order
func (g *Graph) ComputeProcessingOrder() ([]Node, error) {
	order := make([]Node, 0, len(g.nodes))
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(node Node) error
	visit = func(node Node) error {
		if temp[node.GetName()] {
			return fmt.Errorf("cycle detected: %s", node.GetName())
		}
		if visited[node.GetName()] {
			return nil
		}
		temp[node.GetName()] = true

		// Visit parents first to ensure they are processed before this node
		for _, parent := range node.GetParents() {
			if err := visit(parent); err != nil {
				return err
			}
		}

		delete(temp, node.GetName())
		visited[node.GetName()] = true
		order = append(order, node)
		return nil
	}

	// Start with leaf nodes (nodes with no children)
	for _, node := range g.nodes {
		if len(node.GetChildren()) == 0 {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Process any remaining nodes
	for _, node := range g.nodes {
		if !visited[node.GetName()] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

// Removes redundant edges, rejecting cycles before modifying the graph.
func (g *Graph) ComputeTransitiveDependencies() error {
	if _, err := g.ComputeProcessingOrder(); err != nil {
		return err
	}

	for _, node := range g.nodes {
		var newParents []Node
		for _, parent := range node.GetParents() {
			isRedundant := false
			for _, otherParent := range node.GetParents() {
				if otherParent == parent {
					continue
				}

				visited := make(map[string]bool)
				var traverse func(Node)
				traverse = func(n Node) {
					if n == parent {
						isRedundant = true
						return
					}
					for _, p := range n.GetParents() {
						if !visited[p.GetName()] {
							visited[p.GetName()] = true
							traverse(p)
						}
					}
				}
				traverse(otherParent)

				if isRedundant {
					break
				}
			}

			if !isRedundant {
				newParents = append(newParents, parent)
			} else {
				// Remove child relationship from parent
				parentChildren := removeNodeFromSlice(parent.GetChildren(), node)
				parent.SetChildren(parentChildren)
			}
		}
		node.SetParents(newParents)
	}
	return nil
}

// removeNodeFromSlice removes a node from a slice of nodes
func removeNodeFromSlice(nodes []Node, target Node) []Node {
	result := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		if n != target {
			result = append(result, n)
		}
	}
	return result
}

// PrintGraph prints a human-readable representation of the graph structure
func (g *Graph) PrintGraph() {
	fmt.Println("\nGraph Structure:")
	fmt.Println("=====================")

	for name, node := range g.nodes {
		fmt.Printf("\nNode: %s\n", name)
		fmt.Printf("  Parents (%d):\n", len(node.GetParents()))
		for _, parent := range node.GetParents() {
			fmt.Printf("    - %s\n", parent.GetName())
		}

		fmt.Printf("  Children (%d):\n", len(node.GetChildren()))
		for _, child := range node.GetChildren() {
			fmt.Printf("    - %s\n", child.GetName())
		}
	}
	fmt.Println("\n=====================")
}
