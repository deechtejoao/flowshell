package app

import (
	"encoding/base64"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ClipboardWire struct {
	StartNodeID int
	StartPort   int
	EndNodeID   int
	EndPort     int
}

var _ Serializable = &ClipboardWire{}

func (w *ClipboardWire) Serialize(s *Serializer) bool {
	SInt(s, &w.StartNodeID)
	SInt(s, &w.StartPort)
	SInt(s, &w.EndNodeID)
	SInt(s, &w.EndPort)
	return s.Ok()
}

type ClipboardData struct {
	Nodes []*Node
	Wires []*ClipboardWire
	// Groups []*Group // Future: support group copying
}

var _ Serializable = &ClipboardData{}

func (c *ClipboardData) Serialize(s *Serializer) bool {
	// Nodes
	nodeCount := len(c.Nodes)
	SInt(s, &nodeCount)
	if !s.Encode {
		c.Nodes = make([]*Node, nodeCount)
	}
	for i := range nodeCount {
		if !s.Encode {
			c.Nodes[i] = &Node{}
		}
		SThing(s, c.Nodes[i])
	}

	// Wires
	wireCount := len(c.Wires)
	SInt(s, &wireCount)
	if !s.Encode {
		c.Wires = make([]*ClipboardWire, wireCount)
	}
	for i := range wireCount {
		if !s.Encode {
			c.Wires[i] = &ClipboardWire{}
		}
		SThing(s, c.Wires[i])
	}

	return s.Ok()
}

func Copy() {
	if len(SelectedNodes) == 0 {
		return
	}

	data := CopyToData()

	// Serialize
	s := NewEncoder(3)
	if SThing(s, data) {
		str := base64.StdEncoding.EncodeToString(s.Bytes())
		rl.SetClipboardText(str)
	}
}

func CopyToData() *ClipboardData {
	data := &ClipboardData{
		Nodes: []*Node{},
		Wires: []*ClipboardWire{},
	}

	// 1. Collect selected nodes
	for id := range SelectedNodes {
		if n, ok := CurrentGraph.GetNode(id); ok {
			data.Nodes = append(data.Nodes, n)
		}
	}

	// 2. Collect wires between selected nodes
	for _, wire := range CurrentGraph.Wires {
		_, startSelected := SelectedNodes[wire.StartNode.ID]
		_, endSelected := SelectedNodes[wire.EndNode.ID]
		if startSelected && endSelected {
			data.Wires = append(data.Wires, &ClipboardWire{
				StartNodeID: wire.StartNode.ID,
				StartPort:   wire.StartPort,
				EndNodeID:   wire.EndNode.ID,
				EndPort:     wire.EndPort,
			})
		}
	}

	return data
}

func Paste() {
	str := rl.GetClipboardText()
	if str == "" {
		return
	}

	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		// Not a valid base64 string, maybe just text? Ignore.
		return
	}

	s := NewDecoder(bytes)
	var data ClipboardData
	if !SThing(s, &data) {
		// Failed to deserialize
		return
	}

	PasteFromData(&data)
}

func PasteFromData(data *ClipboardData) {
	if len(data.Nodes) == 0 {
		return
	}

	// 1. Prepare ID mapping and offset
	idMap := make(map[int]int)
	offset := V2{X: 20, Y: 20}

	// Clear selection so we can select pasted nodes
	for k := range SelectedNodes {
		delete(SelectedNodes, k)
	}
	selectedNodeID = 0

	// 2. Add nodes
	for _, n := range data.Nodes {
		oldID := n.ID
		CurrentGraph.AddNode(n) // Assigns new ID
		idMap[oldID] = n.ID

		n.Pos = rl.Vector2Add(n.Pos, offset)
		n.Pos = SnapToGrid(n.Pos)

		SelectNode(n.ID, true)
	}

	// 3. Add wires
	for _, w := range data.Wires {
		startID, okStart := idMap[w.StartNodeID]
		endID, okEnd := idMap[w.EndNodeID]

		if okStart && okEnd {
			startNode, ok1 := CurrentGraph.GetNode(startID)
			endNode, ok2 := CurrentGraph.GetNode(endID)
			if ok1 && ok2 {
				CurrentGraph.Wires = append(CurrentGraph.Wires, &Wire{
					StartNode: startNode,
					StartPort: w.StartPort,
					EndNode:   endNode,
					EndPort:   w.EndPort,
				})
			}
		}
	}
}

