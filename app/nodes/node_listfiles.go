package nodes

import (
	"context"
	"errors"
	"fmt"
	"os"

	"path/filepath"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type ListFilesAction struct {
	Dir string
}

func NewListFilesNode(dir string) *core.Node {
	return &core.Node{
		Name: "List Files",

		InputPorts: []core.NodePort{{
			Name: "Directory Path",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Files",
			Type: core.FlowType{Kind: core.FSKindTable, ContainedType: core.FSFile},
		}},

		Action: &ListFilesAction{
			Dir: dir,
		},
	}
}

var _ core.NodeAction = &ListFilesAction{}

func (c *ListFilesAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (c *ListFilesAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("ListFilesUI", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:         core.GROWH,
			ChildAlignment: core.YCENTER,
		},
	}, func() {
		core.PortAnchor(n, false, 0)
		core.UITextBox(clay.IDI("ListFilesDir", n.ID), &c.Dir, core.UITextBoxConfig{
			El:       clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
			Disabled: n.InputIsWired(0),
		})
		core.UISpacer(clay.IDI("ListFilesSpacer", n.ID), core.W2)
		core.UIOutputPort(n, 0)
	})
}

func (c *ListFilesAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)

	go func() {
		var res core.NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		wireDir, hasWire, err := n.GetInputValue(0)
		if err != nil {
			res.Err = err
			return
		}
		dir := util.Tern(hasWire, string(wireDir.BytesValue), c.Dir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			res.Err = err
			return
		}

		var rows [][]core.FlowValueField
		for _, entry := range entries {
			info, err := entry.Info()
			if errors.Is(err, os.ErrNotExist) {
				// This can happen if a file was deleted since the dir was listed. Unlikely but hey.
				continue
			} else if err != nil {
				res.Err = err
				return
			}

			row := []core.FlowValueField{
				{Name: "name", Value: core.NewStringValue(entry.Name())},
				{Name: "path", Value: core.NewStringValue(filepath.Join(dir, entry.Name()))},
				{Name: "type", Value: core.NewStringValue(util.Tern(entry.IsDir(), "dir", "file"))},
				{Name: "size", Value: core.NewInt64Value(info.Size(), core.FSUnitBytes)},
				{Name: "modified", Value: core.NewTimestampValue(info.ModTime())},
			}
			rows = append(rows, row)
		}

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:       &core.FlowType{Kind: core.FSKindTable, ContainedType: core.FSFile},
				TableValue: rows,
			}},
		}
	}()
	return done
}

func (c *ListFilesAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (n *ListFilesAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &n.Dir)
	return s.Ok()
}