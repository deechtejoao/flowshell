package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bvisness/flowshell/clay"
)

func toString(v FlowValue) (string, error) {
	if v.Type.Kind == FSKindBytes {
		return string(v.BytesValue), nil
	}
	return "", fmt.Errorf("expected string/bytes, got %s", v.Type)
}

// GEN:NodeAction
type CopyFileAction struct{}

func NewCopyFileNode() *Node {
	return &Node{
		Name: "Copy File",
		InputPorts: []NodePort{
			{Name: "Source", Type: FlowType{Kind: FSKindAny}},
			{Name: "Destination", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &CopyFileAction{},
	}
}

var _ NodeAction = &CopyFileAction{}

func (a *CopyFileAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *CopyFileAction) UI(n *Node) {
	clay.CLAY(clay.IDI("CopyFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		UIInputPort(n, 0)
		UIInputPort(n, 1)
		UIOutputPort(n, 0)
	})
}

func (a *CopyFileAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *CopyFileAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		srcVal, ok1, err1 := n.GetInputValue(0)
		dstVal, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			done <- NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		srcPath, err := toString(srcVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid source: %w", err)}
			return
		}
		dstPath, err := toString(dstVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid destination: %w", err)}
			return
		}

		// If destination is a directory, append filename
		if info, err := os.Stat(dstPath); err == nil && info.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		sourceFile, err := os.Open(srcPath)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to open source: %w", err)}
			return
		}
		defer func() { _ = sourceFile.Close() }()

		destFile, err := os.Create(dstPath)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to create destination: %w", err)}
			return
		}
		defer func() { _ = destFile.Close() }()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to copy: %w", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(dstPath)}},
		}
	}()

	return done
}

func (a *CopyFileAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type MoveFileAction struct{}

func NewMoveFileNode() *Node {
	return &Node{
		Name: "Move File",
		InputPorts: []NodePort{
			{Name: "Source", Type: FlowType{Kind: FSKindAny}},
			{Name: "Destination", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &MoveFileAction{},
	}
}

var _ NodeAction = &MoveFileAction{}

func (a *MoveFileAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *MoveFileAction) UI(n *Node) {
	clay.CLAY(clay.IDI("MoveFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		UIInputPort(n, 0)
		UIInputPort(n, 1)
		UIOutputPort(n, 0)
	})
}

func (a *MoveFileAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MoveFileAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		srcVal, ok1, err1 := n.GetInputValue(0)
		dstVal, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			done <- NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		srcPath, err := toString(srcVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid source: %w", err)}
			return
		}
		dstPath, err := toString(dstVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid destination: %w", err)}
			return
		}

		// If destination is a directory, append filename
		if info, err := os.Stat(dstPath); err == nil && info.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to move: %w", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(dstPath)}},
		}
	}()

	return done
}

func (a *MoveFileAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type DeleteFileAction struct{}

func NewDeleteFileNode() *Node {
	return &Node{
		Name: "Delete File",
		InputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &DeleteFileAction{},
	}
}

var _ NodeAction = &DeleteFileAction{}

func (a *DeleteFileAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *DeleteFileAction) UI(n *Node) {
	clay.CLAY(clay.IDI("DeleteFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		UIInputPort(n, 0)
		UIOutputPort(n, 0)
	})
}

func (a *DeleteFileAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *DeleteFileAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		pathVal, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}
		if !ok {
			done <- NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		path, err := toString(pathVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid path: %w", err)}
			return
		}

		if err := os.RemoveAll(path); err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to delete: %w", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(path)}},
		}
	}()

	return done
}

func (a *DeleteFileAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type MakeDirAction struct{}

func NewMakeDirNode() *Node {
	return &Node{
		Name: "Make Directory",
		InputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Path", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &MakeDirAction{},
	}
}

var _ NodeAction = &MakeDirAction{}

func (a *MakeDirAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *MakeDirAction) UI(n *Node) {
	clay.CLAY(clay.IDI("MakeDirNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		UIInputPort(n, 0)
		UIOutputPort(n, 0)
	})
}

func (a *MakeDirAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MakeDirAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		pathVal, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}
		if !ok {
			done <- NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		path, err := toString(pathVal)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid path: %w", err)}
			return
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to create directory: %w", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(path)}},
		}
	}()

	return done
}

func (a *MakeDirAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
