package app

import (
	"slices"
)

type Graph struct {
	Nodes      []*Node
	Wires      []*Wire
	NextNodeID int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: []*Node{},
		Wires: []*Wire{},
	}
}

func (g *Graph) AddNode(n *Node) {
	g.NextNodeID++
	n.ID = g.NextNodeID
	n.Graph = g
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) DeleteNode(id int) {
	g.Nodes = slices.DeleteFunc(g.Nodes, func(node *Node) bool { return node.ID == id })
	g.Wires = slices.DeleteFunc(g.Wires, func(wire *Wire) bool { return wire.StartNode.ID == id || wire.EndNode.ID == id })
}

func (g *Graph) GetNode(id int) (*Node, bool) {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return nil, false
}
