package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"path/filepath"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type ListFilesAction struct {
	Dir string
}

func NewListFilesNode(dir string) *Node {
	return &Node{
		Name: "List Files",

		InputPorts: []NodePort{{
			Name: "Directory Path",
			Type: FlowType{Kind: FSKindBytes},
		}},
		OutputPorts: []NodePort{{
			Name: "Files",
			Type: FlowType{Kind: FSKindTable, ContainedType: FSFile},
		}},

		Action: &ListFilesAction{
			Dir: dir,
		},
	}
}

var _ NodeAction = &ListFilesAction{}

func (c *ListFilesAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (c *ListFilesAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		PortAnchor(n, false, 0)
		UITextBox(clay.IDI("ListFilesDir", n.ID), &c.Dir, UITextBoxConfig{
			El:       clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			Disabled: n.InputIsWired(0),
		})
		UISpacer(clay.AUTO_ID, W2)
		UIOutputPort(n, 0)
	})
}

func (c *ListFilesAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	go func() {
		var res NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
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

		var rows [][]FlowValueField
		for _, entry := range entries {
			info, err := entry.Info()
			if errors.Is(err, os.ErrNotExist) {
				// This can happen if a file was deleted since the dir was listed. Unlikely but hey.
				continue
			} else if err != nil {
				res.Err = err
				return
			}

			row := []FlowValueField{
				{Name: "name", Value: NewStringValue(entry.Name())},
				{Name: "path", Value: NewStringValue(filepath.Join(dir, entry.Name()))},
				{Name: "type", Value: NewStringValue(util.Tern(entry.IsDir(), "dir", "file"))},
				{Name: "size", Value: NewInt64Value(info.Size(), FSUnitBytes)},
				{Name: "modified", Value: NewTimestampValue(info.ModTime())},
			}
			rows = append(rows, row)
		}

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:       &FlowType{Kind: FSKindTable, ContainedType: FSFile},
				TableValue: rows,
			}},
		}
	}()
	return done
}

func (c *ListFilesAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (n *ListFilesAction) Serialize(s *Serializer) bool {
	SStr(s, &n.Dir)
	return s.Ok()
}
