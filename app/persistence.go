package app

import (
	"fmt"
	"os"
)

func SaveGraph(path string) error {
	s := NewEncoder(2)

	// Nodes
	nodeCount := len(nodes)
	SInt(s, &nodeCount)
	for _, n := range nodes {
		SThing(s, n)
	}

	// Wires
	wireCount := len(wires)
	SInt(s, &wireCount)
	for _, w := range wires {
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

func LoadGraph(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	s := NewDecoder(data)

	// Read into temporary variables first to avoid destroying state on failure
	var newNodes []*Node
	var newWires []*Wire
	var newNodeID int

	// Nodes
	var nodeCount int
	if !SInt(s, &nodeCount) {
		return fmt.Errorf("failed to read node count: %v", s.Errs)
	}

	nodeMap := make(map[int]*Node)
	maxID := 0

	for range nodeCount {
		n := &Node{}
		if !SThing(s, n) {
			return fmt.Errorf("failed to read node: %v", s.Errs)
		}
		newNodes = append(newNodes, n)
		nodeMap[n.ID] = n
		if n.ID > maxID {
			maxID = n.ID
		}
	}
	newNodeID = maxID // Ensure next new node gets a unique ID

	// Wires
	var wireCount int
	if !SInt(s, &wireCount) {
		return fmt.Errorf("failed to read wire count: %v", s.Errs)
	}

	for range wireCount {
		var startNodeID, startPort, endNodeID, endPort int
		SInt(s, &startNodeID)
		SInt(s, &startPort)
		SInt(s, &endNodeID)
		SInt(s, &endPort)

		if !s.Ok() {
			return fmt.Errorf("failed to read wire: %v", s.Errs)
		}

		startNode, ok1 := nodeMap[startNodeID]
		endNode, ok2 := nodeMap[endNodeID]

		if ok1 && ok2 {
			newWires = append(newWires, &Wire{
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
		return fmt.Errorf("deserialization failed: %v", s.Errs)
	}

	// Success! Swap in the new state
	nodes = newNodes
	wires = newWires
	nodeID = newNodeID

	return nil
}
