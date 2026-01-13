package app

import (
	"slices"

	"github.com/bvisness/flowshell/clay"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Graph struct {
	Nodes      []*Node
	Wires      []*Wire
	Groups     []*Group
	NextNodeID int
	NextGroupID int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:  []*Node{},
		Wires:  []*Wire{},
		Groups: []*Group{},
	}
}

func (g *Graph) AddNode(n *Node) {
	g.NextNodeID++
	n.ID = g.NextNodeID
	n.Graph = g
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) DeleteNode(id int) {
	// When deleting a node, we must invalidate any nodes that were connected to it,
	// because their inputs/outputs have changed.
	for _, wire := range g.Wires {
		if wire.StartNode.ID == id {
			wire.EndNode.ClearResult()
		}
	}

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

type Group struct {
	ID    int
	Title string
	Rect  rl.Rectangle
	Color clay.Color
}

var _ Serializable = &Group{}

func (g *Group) Serialize(s *Serializer) bool {
	SInt(s, &g.ID)
	SStr(s, &g.Title)
	SFloat(s, &g.Rect.X)
	SFloat(s, &g.Rect.Y)
	SFloat(s, &g.Rect.Width)
	SFloat(s, &g.Rect.Height)
	SFloat(s, &g.Color.R)
	SFloat(s, &g.Color.G)
	SFloat(s, &g.Color.B)
	SFloat(s, &g.Color.A)
	return s.Ok()
}

func (g *Graph) AddGroup(grp *Group) {
	g.NextGroupID++
	grp.ID = g.NextGroupID
	g.Groups = append(g.Groups, grp)
}

func (g *Graph) DeleteGroup(id int) {
	g.Groups = slices.DeleteFunc(g.Groups, func(grp *Group) bool { return grp.ID == id })
}

func (g *Graph) GetGroup(id int) (*Group, bool) {
	for _, grp := range g.Groups {
		if grp.ID == id {
			return grp, true
		}
	}
	return nil, false
}
