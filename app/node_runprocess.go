package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type RunProcessAction struct {
	CmdString string
}

func NewRunProcessNode(cmd string) *Node {
	return &Node{
		ID:   NewNodeID(),
		Name: "Run Process",

		InputPorts: nil,
		OutputPorts: []NodePort{
			{
				Name: "Stdout",
				Type: FlowType{Kind: FSKindBytes},
			},
			{
				Name: "Stderr",
				Type: FlowType{Kind: FSKindBytes},
			},
			{
				Name: "Combined Stdout/Stderr",
				Type: FlowType{Kind: FSKindBytes},
			},
			{
				Name: "Exit Code",
				Type: FlowType{Kind: FSKindInt64},
			},
		},

		Action: &RunProcessAction{
			CmdString: cmd,
		},
	}
}

var _ NodeAction = &RunProcessAction{}

func (c *RunProcessAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (c *RunProcessAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		UITextBox(clay.IDI("RunProcessCmd", n.ID), &c.CmdString, UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
		})

		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          GROWH,
				ChildAlignment:  XRIGHT,
			},
		}, func() {
			UIOutputPort(n, 0)
			UIOutputPort(n, 1)
			UIOutputPort(n, 2)
			UIOutputPort(n, 3)
		})
	})
}

func (c *RunProcessAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	pieces := parseCommand(c.CmdString)
	if len(pieces) == 0 {
		go func() {
			done <- NodeActionResult{Err: nil} // Or error?
			close(done)
		}()
		return done
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(cmdCtx, pieces[0], pieces[1:]...)

	var stdout, stderr, combined []byte
	var exitCode int64
	var runErr error
	var mu sync.Mutex

	cmd.Stdout = &multiSliceWriter{
		mu: &mu,
		a:  &stdout,
		b:  &combined,
	}
	cmd.Stderr = &multiSliceWriter{
		mu: &mu,
		a:  &stderr,
		b:  &combined,
	}

	go func() {
		var res NodeActionResult
		defer close(done)
		defer cancel()
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

		runErr = cmd.Run()
		if runErr != nil {
			if exitErr, ok := runErr.(*exec.ExitError); ok {
				exitCode = int64(exitErr.ExitCode())
				runErr = nil // We captured the exit code, so the node execution is "successful"
			} else {
				exitCode = -1 // Unknown error or signal
			}
		} else {
			exitCode = 0
		}

		res = NodeActionResult{
			Err: runErr,
			Outputs: []FlowValue{
				{
					Type:       &FlowType{Kind: FSKindBytes},
					BytesValue: stdout,
				},
				{
					Type:       &FlowType{Kind: FSKindBytes},
					BytesValue: stderr,
				},
				{
					Type:       &FlowType{Kind: FSKindBytes},
					BytesValue: combined,
				},
				{
					Type:       &FlowType{Kind: FSKindInt64},
					Int64Value: exitCode,
				},
			},
		}
	}()
	return done
}

func (c *RunProcessAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *RunProcessAction) Serialize(s *Serializer) bool {
	SStr(s, &c.CmdString)
	return s.Ok()
}

func parseCommand(cmd string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range cmd {
		switch {
		case inQuote:
			if r == quoteChar {
				inQuote = false
			} else {
				current.WriteRune(r)
			}
		case r == '"' || r == '\'':
			inQuote = true
			quoteChar = r
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
