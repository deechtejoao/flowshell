package core

import (
	"fmt"
)

type GraphBuilder struct {
	Graph *Graph
	Nodes map[string]*NodeBuilder
}

func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		Graph: NewGraph(),
		Nodes: make(map[string]*NodeBuilder),
	}
}

func (gb *GraphBuilder) Add(node *Node) *NodeBuilder {
	gb.Graph.AddNode(node)
	nb := &NodeBuilder{
		Builder: gb,
		Node:    node,
	}
	// Default key is name, but unique handling might be needed if duplicates exist.
	// For now, let's just key by pointer string or require explicit alias if needed?
	// The map is mainly for internal lookups if we added a "Get" method.
	// For fluent API, we return the NodeBuilder.
	return nb
}

type NodeBuilder struct {
	Builder *GraphBuilder
	Node    *Node
}

func (nb *NodeBuilder) SetPosition(x, y float32) *NodeBuilder {
	nb.Node.Pos = V2{X: x, Y: y}
	return nb
}

// ConnectTo Connects an output port of this node to an input port of another node.
// Returns the *NodeBuilder of the *destination* node to allow chaining?
// Or returns the GraphBuilder? Or the source NodeBuilder?
// "Fluency" usually implies chaining.
// A -> B -> C
// nodeA.Connect("Out", nodeB, "In").Connect("Out", nodeC, "In") ?
// But Connect returns NodeBuilder? Which one?
// If it returns 'dst', then A.Connect(B).Connect(C) implies B->C.
// That works for linear chains.
func (nb *NodeBuilder) Connect(outPortName string, dst *NodeBuilder, inPortName string) *NodeBuilder {
	// Find ports
	outPortIdx := -1
	for i, p := range nb.Node.OutputPorts {
		if p.Name == outPortName {
			outPortIdx = i
			break
		}
	}
	if outPortIdx == -1 {
		panic(fmt.Sprintf("Node %s has no output port named %s", nb.Node.Name, outPortName))
	}

	inPortIdx := -1
	for i, p := range dst.Node.InputPorts {
		if p.Name == inPortName {
			inPortIdx = i
			break
		}
	}
	if inPortIdx == -1 {
		panic(fmt.Sprintf("Node %s has no input port named %s", dst.Node.Name, inPortName))
	}

	nb.Builder.Graph.AddWire(nb.Node, outPortIdx, dst.Node, inPortIdx)
	return dst // Return destination to allow chaining: A.Connect(B).Connect(C) -> A->B->C
}

// Named alias for Connect to make chains clearer
func (nb *NodeBuilder) To(dst *NodeBuilder) *NodeBuilder {
	// Defaut connection?
	// Most nodes have 1 Main Input and 1 Main Output.
	// We can try to guess?
	// Safe bet: First Output to First Input.
	if len(nb.Node.OutputPorts) == 0 {
		panic(fmt.Sprintf("Node %s has no output ports", nb.Node.Name))
	}
	if len(dst.Node.InputPorts) == 0 {
		panic(fmt.Sprintf("Node %s has no input ports", dst.Node.Name))
	}
	nb.Builder.Graph.AddWire(nb.Node, 0, dst.Node, 0)
	return dst
}
