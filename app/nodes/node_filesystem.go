package nodes

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

func toString(v core.FlowValue) (string, error) {
	if v.Type.Kind == core.FSKindBytes {
		return string(v.BytesValue), nil
	}
	return "", fmt.Errorf("expected string/bytes, got %s", v.Type)
}

// GEN:NodeAction
type CopyFileAction struct{}

func NewCopyFileNode() *core.Node {
	return &core.Node{
		Name: "Copy File",
		InputPorts: []core.NodePort{
			{Name: "Source", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "Destination", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &CopyFileAction{},
	}
}

var _ core.NodeAction = &CopyFileAction{}

func (a *CopyFileAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *CopyFileAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("CopyFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		core.UIInputPort(n, 0)
		core.UIInputPort(n, 1)
		core.UIOutputPort(n, 0)
	})
}

func (a *CopyFileAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *CopyFileAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		srcVal, ok1, err1 := n.GetInputValue(0)
		dstVal, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- core.NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- core.NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			done <- core.NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		srcPath, err := toString(srcVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid source: %w", err)}
			return
		}
		dstPath, err := toString(dstVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid destination: %w", err)}
			return
		}

		// If destination is a directory, append filename
		if info, err := os.Stat(dstPath); err == nil && info.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		sourceFile, err := os.Open(srcPath)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to open source: %w", err)}
			return
		}
		defer func() { _ = sourceFile.Close() }()

		destFile, err := os.Create(dstPath)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to create destination: %w", err)}
			return
		}
		defer func() { _ = destFile.Close() }()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to copy: %w", err)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(dstPath)}},
		}
	}()

	return done
}

func (a *CopyFileAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type MoveFileAction struct{}

func NewMoveFileNode() *core.Node {
	return &core.Node{
		Name: "Move File",
		InputPorts: []core.NodePort{
			{Name: "Source", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "Destination", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &MoveFileAction{},
	}
}

var _ core.NodeAction = &MoveFileAction{}

func (a *MoveFileAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *MoveFileAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("MoveFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		core.UIInputPort(n, 0)
		core.UIInputPort(n, 1)
		core.UIOutputPort(n, 0)
	})
}

func (a *MoveFileAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MoveFileAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		srcVal, ok1, err1 := n.GetInputValue(0)
		dstVal, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- core.NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- core.NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			done <- core.NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		srcPath, err := toString(srcVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid source: %w", err)}
			return
		}
		dstPath, err := toString(dstVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid destination: %w", err)}
			return
		}

		// If destination is a directory, append filename
		if info, err := os.Stat(dstPath); err == nil && info.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to move: %w", err)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(dstPath)}},
		}
	}()

	return done
}

func (a *MoveFileAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type DeleteFileAction struct{}

func NewDeleteFileNode() *core.Node {
	return &core.Node{
		Name: "Delete File",
		InputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &DeleteFileAction{},
	}
}

var _ core.NodeAction = &DeleteFileAction{}

func (a *DeleteFileAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *DeleteFileAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("DeleteFileNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		core.UIInputPort(n, 0)
		core.UIOutputPort(n, 0)
	})
}

func (a *DeleteFileAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *DeleteFileAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		pathVal, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}
		if !ok {
			done <- core.NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		path, err := toString(pathVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid path: %w", err)}
			return
		}

		if err := os.RemoveAll(path); err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to delete: %w", err)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(path)}},
		}
	}()

	return done
}

func (a *DeleteFileAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type MakeDirAction struct{}

func NewMakeDirNode() *core.Node {
	return &core.Node{
		Name: "Make Directory",
		InputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Path", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &MakeDirAction{},
	}
}

var _ core.NodeAction = &MakeDirAction{}

func (a *MakeDirAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *MakeDirAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("MakeDirNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		core.UIInputPort(n, 0)
		core.UIOutputPort(n, 0)
	})
}

func (a *MakeDirAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MakeDirAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		pathVal, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}
		if !ok {
			done <- core.NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		path, err := toString(pathVal)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("invalid path: %w", err)}
			return
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to create directory: %w", err)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(path)}},
		}
	}()

	return done
}

func (a *MakeDirAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}
