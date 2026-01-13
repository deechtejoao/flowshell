package app

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type RunProcessAction struct {
	CmdString    string
	UseShell     bool
	UseStdin     bool
	StreamOutput bool
}

func NewRunProcessNode(cmd string) *Node {
	return &Node{
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

	if c.UseStdin {
		if len(n.InputPorts) == 0 {
			n.InputPorts = []NodePort{{
				Name: "Stdin",
				Type: FlowType{Kind: FSKindAny}, // Accept Bytes or Stream
			}}
		}
	} else {
		n.InputPorts = nil
	}

	outKind := FSKindBytes
	if c.StreamOutput {
		outKind = FSKindStream
	}

	n.OutputPorts[0].Type = FlowType{Kind: outKind} // Stdout
	n.OutputPorts[1].Type = FlowType{Kind: outKind} // Stderr
	n.OutputPorts[2].Type = FlowType{Kind: outKind} // Combined
	// ExitCode remains Int64
}

func (c *RunProcessAction) UI(n *Node) {
	clay.CLAY(clay.IDI("RunProcessUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		UITextBox(clay.IDI("RunProcessCmd", n.ID), &c.CmdString, UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
		})

		UIButton(clay.IDI("RunProcessShellBtn", n.ID), UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: S2, ChildAlignment: YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.UseShell = !c.UseShell
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessShellCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: WH(16, 16)},
				Border:          clay.B{Width: BA, Color: White},
				BackgroundColor: util.Tern(c.UseShell, Blue, clay.Color{}),
			})
			clay.TEXT("Use Shell", clay.TextElementConfig{TextColor: White})
		})

		UIButton(clay.IDI("RunProcessStdinBtn", n.ID), UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: S2, ChildAlignment: YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.UseStdin = !c.UseStdin
				c.UpdateAndValidate(n)
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessStdinCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: WH(16, 16)},
				Border:          clay.B{Width: BA, Color: White},
				BackgroundColor: util.Tern(c.UseStdin, Blue, clay.Color{}),
			})
			clay.TEXT("Use Stdin", clay.TextElementConfig{TextColor: White})
		})

		UIButton(clay.IDI("RunProcessStreamBtn", n.ID), UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: S2, ChildAlignment: YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.StreamOutput = !c.StreamOutput
				c.UpdateAndValidate(n)
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessStreamCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: WH(16, 16)},
				Border:          clay.B{Width: BA, Color: White},
				BackgroundColor: util.Tern(c.StreamOutput, Blue, clay.Color{}),
			})
			clay.TEXT("Stream Output", clay.TextElementConfig{TextColor: White})
		})

		clay.CLAY(clay.IDI("RunProcessOutputs", n.ID), clay.EL{
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          GROWH,
				ChildAlignment:  XRIGHT,
			},
		}, func() {
			if c.UseStdin {
				UIInputPort(n, 0)
			}
			UIOutputPort(n, 0)
			UIOutputPort(n, 1)
			UIOutputPort(n, 2)
			UIOutputPort(n, 3)
		})
	})
}

func (c *RunProcessAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	cmdCtx, cancel := context.WithCancel(ctx)
	var cmd *exec.Cmd

	if c.UseShell {
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(cmdCtx, "powershell", "-Command", c.CmdString)
		} else {
			cmd = exec.CommandContext(cmdCtx, "sh", "-c", c.CmdString)
		}
	} else {
		pieces := parseCommand(c.CmdString)
		if len(pieces) == 0 {
			go func() {
				done <- NodeActionResult{Err: nil} // Or error?
				close(done)
			}()
			cancel()
			return done
		}
		cmd = exec.CommandContext(cmdCtx, pieces[0], pieces[1:]...)
	}

	// Handle Stdin
	if c.UseStdin {
		input, ok, err := n.GetInputValue(0)
		if err != nil {
			go func() {
				done <- NodeActionResult{Err: err}
				close(done)
			}()
			cancel()
			return done
		}
		if ok {
			switch input.Type.Kind {
			case FSKindBytes:
				cmd.Stdin = strings.NewReader(string(input.BytesValue))
			case FSKindStream:
				if input.StreamValue != nil {
					cmd.Stdin = input.StreamValue
				}
			}
		}
	}

	if c.StreamOutput {
		// Setup pipes
		stdoutPR, stdoutPW := io.Pipe()
		stderrPR, stderrPW := io.Pipe()
		combinedPR, combinedPW := io.Pipe()

		cmd.Stdout = io.MultiWriter(stdoutPW, combinedPW)
		cmd.Stderr = io.MultiWriter(stderrPW, combinedPW)

		// Start command
		if err := cmd.Start(); err != nil {
			// Close all pipe ends to prevent leaks
			_ = stdoutPR.Close()
			_ = stdoutPW.Close()
			_ = stderrPR.Close()
			_ = stderrPW.Close()
			_ = combinedPR.Close()
			_ = combinedPW.Close()

			go func() {
				done <- NodeActionResult{Err: err}
				close(done)
			}()
			cancel()
			return done
		}

		// Return streams immediately
		go func() {
			done <- NodeActionResult{
				Outputs: []FlowValue{
					{Type: &FlowType{Kind: FSKindStream}, StreamValue: stdoutPR},
					{Type: &FlowType{Kind: FSKindStream}, StreamValue: stderrPR},
					{Type: &FlowType{Kind: FSKindStream}, StreamValue: combinedPR},
					{Type: &FlowType{Kind: FSKindInt64}, Int64Value: 0}, // Exit code unavailable
				},
			}
			close(done)

			// Wait for command in background to close pipes
			go func() {
				defer cancel() // Ensure context is cancelled eventually
				_ = cmd.Wait()
				_ = stdoutPW.Close()
				_ = stderrPW.Close()
				_ = combinedPW.Close()
			}()
		}()

		return done
	}

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
		if ctx.Err() != nil {
			res.Err = ctx.Err()
			return
		}

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

	if !s.Encode && s.Buf.Len() == 0 {
		c.UseShell = false
		c.UseStdin = false
		c.StreamOutput = false
		return s.Ok()
	}

	SBool(s, &c.UseShell)

	if !s.Encode && s.Buf.Len() == 0 {
		c.UseStdin = false
		c.StreamOutput = false
		return s.Ok()
	}

	SBool(s, &c.UseStdin)
	SBool(s, &c.StreamOutput)
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
