package app

import (
	"bytes"
	"fmt"
)

// HistoryManager handles Undo/Redo by storing serialized snapshots of the graph.
type HistoryManager struct {
	snapshots [][]byte
	pointer   int // points to the current snapshot
}

// NewHistoryManager creates a new history manager.
// initialGraph is the starting state.
func NewHistoryManager(initialGraph *Graph) *HistoryManager {
	hm := &HistoryManager{
		snapshots: make([][]byte, 0),
		pointer:   -1,
	}
	hm.Push(initialGraph)
	return hm
}

// Push adds a new snapshot to the history.
// It truncates any redo history.
func (hm *HistoryManager) Push(g *Graph) {
	data, err := SerializeGraph(g)
	if err != nil {
		fmt.Printf("History Push failed: %v\n", err)
		return
	}

	// Deduplication: Don't push if identical to current snapshot
	if hm.pointer >= 0 && hm.pointer < len(hm.snapshots) {
		if bytes.Equal(hm.snapshots[hm.pointer], data) {
			return
		}
	}

	// Truncate redo history
	if hm.pointer < len(hm.snapshots)-1 {
		hm.snapshots = hm.snapshots[:hm.pointer+1]
	}

	hm.snapshots = append(hm.snapshots, data)
	hm.pointer++

	// Limit history size (optional, e.g. 50 steps)
	if len(hm.snapshots) > 50 {
		hm.snapshots = hm.snapshots[1:]
		hm.pointer--
	}
}

// Undo reverts to the previous snapshot.
// Returns the new graph state, or nil if cannot undo.
func (hm *HistoryManager) Undo() *Graph {
	if hm.pointer <= 0 {
		return nil
	}

	hm.pointer--
	data := hm.snapshots[hm.pointer]

	g, err := DeserializeGraph(data)
	if err != nil {
		fmt.Printf("History Undo failed: %v\n", err)
		return nil
	}
	return g
}

// Redo advances to the next snapshot.
// Returns the new graph state, or nil if cannot redo.
func (hm *HistoryManager) Redo() *Graph {
	if hm.pointer >= len(hm.snapshots)-1 {
		return nil
	}

	hm.pointer++
	data := hm.snapshots[hm.pointer]

	g, err := DeserializeGraph(data)
	if err != nil {
		fmt.Printf("History Redo failed: %v\n", err)
		return nil
	}
	return g
}

// CanUndo returns true if undo is possible.
func (hm *HistoryManager) CanUndo() bool {
	return hm.pointer > 0
}

// CanRedo returns true if redo is possible.
func (hm *HistoryManager) CanRedo() bool {
	return hm.pointer < len(hm.snapshots)-1
}
