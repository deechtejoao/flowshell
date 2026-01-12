package app

import (
	"fmt"
	"os"
)

func SaveGraph(path string, g *Graph) error {
	s := NewEncoder(2)

	// Nodes
	nodeCount := len(g.Nodes)
	SInt(s, &nodeCount)
	for _, n := range g.Nodes {
		SThing(s, n)
	}

	// Wires
	wireCount := len(g.Wires)
	SInt(s, &wireCount)
	for _, w := range g.Wires {
		SInt(s, &w.StartNode.ID)
		SInt(s, &w.StartPort)
		SInt(s, &w.EndNode.ID)
		SInt(s, &w.EndPort)
	}

	if !s.Ok() {
		return fmt.Errorf("serialization failed: %v", s.Errs)
	}

	return os.WriteFile(path, s.Bytes(), 0644)
}

func LoadGraph(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := NewDecoder(data)

	// Read into temporary variables first to avoid destroying state on failure
	g := NewGraph()

	// Nodes
	var nodeCount int
	if !SInt(s, &nodeCount) {
		return nil, fmt.Errorf("failed to read node count: %v", s.Errs)
	}

	nodeMap := make(map[int]*Node)
	maxID := 0

	for range nodeCount {
		n := &Node{}
		if !SThing(s, n) {
			return nil, fmt.Errorf("failed to read node: %v", s.Errs)
		}
		g.AddNode(n)
		nodeMap[n.ID] = n
		if n.ID > maxID {
			maxID = n.ID
		}
	}
	g.NextNodeID = maxID

	// Wires
	var wireCount int
	if !SInt(s, &wireCount) {
		return nil, fmt.Errorf("failed to read wire count: %v", s.Errs)
	}

	for range wireCount {
		var startNodeID, startPort, endNodeID, endPort int
		SInt(s, &startNodeID)
		SInt(s, &startPort)
		SInt(s, &endNodeID)
		SInt(s, &endPort)

		if !s.Ok() {
			return nil, fmt.Errorf("failed to read wire: %v", s.Errs)
		}

		startNode, ok1 := nodeMap[startNodeID]
		endNode, ok2 := nodeMap[endNodeID]

		if ok1 && ok2 {
			g.Wires = append(g.Wires, &Wire{
				StartNode: startNode,
				StartPort: startPort,
				EndNode:   endNode,
				EndPort:   endPort,
			})
		} else {
			fmt.Printf("Warning: skipping wire with missing nodes (%d -> %d)\n", startNodeID, endNodeID)
		}
	}

	if !s.Ok() {
		return nil, fmt.Errorf("deserialization failed: %v", s.Errs)
	}

	return g, nil
}
