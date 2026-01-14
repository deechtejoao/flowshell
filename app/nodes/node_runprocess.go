package nodes

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/bvisness/flowshell/app/core"
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

func NewRunProcessNode(cmd string) *core.Node {
	return &core.Node{
		Name: "Run Process",

		InputPorts: nil,
		OutputPorts: []core.NodePort{
			{
				Name: "Stdout",
				Type: core.FlowType{Kind: core.FSKindBytes},
			},
			{
				Name: "Stderr",
				Type: core.FlowType{Kind: core.FSKindBytes},
			},
			{
				Name: "Combined Stdout/Stderr",
				Type: core.FlowType{Kind: core.FSKindBytes},
			},
			{
				Name: "Exit Code",
				Type: core.FlowType{Kind: core.FSKindInt64},
			},
		},

		Action: &RunProcessAction{
			CmdString: cmd,
		},
	}
}

var _ core.NodeAction = &RunProcessAction{}

func (c *RunProcessAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	if c.UseStdin {
		if len(n.InputPorts) == 0 {
			n.InputPorts = []core.NodePort{{
				Name: "Stdin",
				Type: core.FlowType{Kind: core.FSKindAny}, // Accept Bytes or Stream
			}}
		}
	} else {
		n.InputPorts = nil
	}

	outKind := core.FSKindBytes
	if c.StreamOutput {
		outKind = core.FSKindStream
	}

	n.OutputPorts[0].Type = core.FlowType{Kind: outKind} // Stdout
	n.OutputPorts[1].Type = core.FlowType{Kind: outKind} // Stderr
	n.OutputPorts[2].Type = core.FlowType{Kind: outKind} // Combined
	// ExitCode remains Int64
}

func (c *RunProcessAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("RunProcessUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		core.UITextBox(clay.IDI("RunProcessCmd", n.ID), &c.CmdString, core.UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
		})

		core.UIButton(clay.IDI("RunProcessShellBtn", n.ID), core.UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: core.S2, ChildAlignment: core.YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.UseShell = !c.UseShell
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessShellCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: core.WH(16, 16)},
				Border:          clay.B{Width: core.BA, Color: core.White},
				BackgroundColor: util.Tern(c.UseShell, core.Blue, clay.Color{}),
			})
			clay.TEXT("Use Shell", clay.TextElementConfig{TextColor: core.White})
		})

		core.UIButton(clay.IDI("RunProcessStdinBtn", n.ID), core.UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: core.S2, ChildAlignment: core.YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.UseStdin = !c.UseStdin
				c.UpdateAndValidate(n)
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessStdinCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: core.WH(16, 16)},
				Border:          clay.B{Width: core.BA, Color: core.White},
				BackgroundColor: util.Tern(c.UseStdin, core.Blue, clay.Color{}),
			})
			clay.TEXT("Use Stdin", clay.TextElementConfig{TextColor: core.White})
		})

		core.UIButton(clay.IDI("RunProcessStreamBtn", n.ID), core.UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{ChildGap: core.S2, ChildAlignment: core.YCENTER},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				c.StreamOutput = !c.StreamOutput
				c.UpdateAndValidate(n)
			},
		}, func() {
			clay.CLAY(clay.IDI("RunProcessStreamCheck", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: core.WH(16, 16)},
				Border:          clay.B{Width: core.BA, Color: core.White},
				BackgroundColor: util.Tern(c.StreamOutput, core.Blue, clay.Color{}),
			})
			clay.TEXT("Stream Output", clay.TextElementConfig{TextColor: core.White})
		})

		clay.CLAY(clay.IDI("RunProcessOutputs", n.ID), clay.EL{
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          core.GROWH,
				ChildAlignment:  core.XRIGHT,
			},
		}, func() {
			if c.UseStdin {
				core.UIInputPort(n, 0)
			}
			core.UIOutputPort(n, 0)
			core.UIOutputPort(n, 1)
			core.UIOutputPort(n, 2)
			core.UIOutputPort(n, 3)
		})
	})
}

func (c *RunProcessAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)

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
				done <- core.NodeActionResult{Err: nil} // Or error?
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
				done <- core.NodeActionResult{Err: err}
				close(done)
			}()
			cancel()
			return done
		}
		if ok {
			switch input.Type.Kind {
			case core.FSKindBytes:
				cmd.Stdin = strings.NewReader(string(input.BytesValue))
			case core.FSKindStream:
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
				done <- core.NodeActionResult{Err: err}
				close(done)
			}()
			cancel()
			return done
		}

		// Return streams immediately
		go func() {
			done <- core.NodeActionResult{
				Outputs: []core.FlowValue{
					{Type: &core.FlowType{Kind: core.FSKindStream}, StreamValue: stdoutPR},
					{Type: &core.FlowType{Kind: core.FSKindStream}, StreamValue: stderrPR},
					{Type: &core.FlowType{Kind: core.FSKindStream}, StreamValue: combinedPR},
					{Type: &core.FlowType{Kind: core.FSKindInt64}, Int64Value: 0}, // Exit code unavailable
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
		var res core.NodeActionResult
		defer close(done)
		defer cancel()
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

		res = core.NodeActionResult{
			Err: runErr,
			Outputs: []core.FlowValue{
				{
					Type:       &core.FlowType{Kind: core.FSKindBytes},
					BytesValue: stdout,
				},
				{
					Type:       &core.FlowType{Kind: core.FSKindBytes},
					BytesValue: stderr,
				},
				{
					Type:       &core.FlowType{Kind: core.FSKindBytes},
					BytesValue: combined,
				},
				{
					Type:       &core.FlowType{Kind: core.FSKindInt64},
					Int64Value: exitCode,
				},
			},
		}
	}()
	return done
}

func (c *RunProcessAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *RunProcessAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.CmdString)

	if !s.Encode && s.Buf.Len() == 0 {
		c.UseShell = false
		c.UseStdin = false
		c.StreamOutput = false
		return s.Ok()
	}

	core.SBool(s, &c.UseShell)

	if !s.Encode && s.Buf.Len() == 0 {
		c.UseStdin = false
		c.StreamOutput = false
		return s.Ok()
	}

	core.SBool(s, &c.UseStdin)
	core.SBool(s, &c.StreamOutput)
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

type multiSliceWriter struct {
	mu *sync.Mutex
	a  *[]byte
	b  *[]byte
}

func (w *multiSliceWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	*w.a = append(*w.a, p...)
	*w.b = append(*w.b, p...)
	return len(p), nil
}
