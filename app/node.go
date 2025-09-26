package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type V2 = rl.Vector2

type Node struct {
	ID   int
	Pos  V2
	Name string

	InputPorts  []NodePort
	OutputPorts []NodePort

	Action NodeAction

	Running bool
}

type NodePort struct {
	Name string
	Type FlowType
}

func (n *Node) Run() {
	if n.Running {
		return
	}
	n.Running = true

	done := n.Action.Run(n)
	go func() {
		<-done
		n.Running = false
	}()
}

type NodeAction interface {
	UI(n *Node)
	Run(n *Node) (done <-chan struct{})
	Result() NodeActionResult
}

type NodeActionResult struct {
	Outputs []FlowValue
	Err     error
}

var nodeID = 0

func NewNodeID() int {
	nodeID++
	return nodeID
}

// --------------------------------
// Run Process

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
		},

		Action: &RunProcessAction{
			CmdString: cmd,
		},
	}
}

type RunProcessAction struct {
	CmdString string

	state             RunProcessActionRuntimeState
	outputStreamMutex sync.Mutex
}

// The state that gets reset every time you run a command
type RunProcessActionRuntimeState struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc

	stdout   []byte
	stderr   []byte
	combined []byte

	err      error
	exitCode int
}

func (c *RunProcessAction) UI(n *Node) {
	UITextBox(clay.AUTO_ID, &c.CmdString, clay.EL{Layout: clay.LAY{Sizing: GROWH}})
}

func (c *RunProcessAction) Run(n *Node) <-chan struct{} {
	pieces := strings.Split(c.CmdString, " ")
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, pieces[0], pieces[1:]...)

	done := make(chan struct{})

	c.state = RunProcessActionRuntimeState{
		cmd:    cmd,
		cancel: cancel,
	}

	cmd.Stdout = &multiSliceWriter{
		mu: &c.outputStreamMutex,
		a:  &c.state.stdout,
		b:  &c.state.combined,
	}
	cmd.Stderr = &multiSliceWriter{
		mu: &c.outputStreamMutex,
		a:  &c.state.stderr,
		b:  &c.state.combined,
	}

	go func() {
		defer func() { done <- struct{}{} }()

		c.state.err = c.state.cmd.Run()
		if c.state.err != nil {
			// TODO: Extract exit code
		}
	}()

	return done
}

func (c *RunProcessAction) Result() NodeActionResult {
	return NodeActionResult{
		Err: c.state.err,
		Outputs: []FlowValue{
			{
				Type:       &FlowType{Kind: FSKindBytes},
				BytesValue: c.state.stdout,
			},
			{
				Type:       &FlowType{Kind: FSKindBytes},
				BytesValue: c.state.stderr,
			},
			{
				Type:       &FlowType{Kind: FSKindBytes},
				BytesValue: c.state.combined,
			},
		},
	}
}

// --------------------------------
// List Files

func NewListFilesNode(dir string) *Node {
	return &Node{
		ID:   NewNodeID(),
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

type ListFilesAction struct {
	Dir string

	rows [][]FlowValueField
	err  error
}

func (c *ListFilesAction) UI(n *Node) {
	UITextBox(clay.ID("Cmd"), &c.Dir, clay.EL{Layout: clay.LAY{Sizing: GROWH}})
}

func (c *ListFilesAction) Run(n *Node) <-chan struct{} {
	done := make(chan struct{})

	c.rows = nil
	c.err = nil

	go func() {
		defer func() { done <- struct{}{} }()

		entries, err := os.ReadDir(c.Dir)
		if err != nil {
			c.err = err
			return
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if errors.Is(err, os.ErrNotExist) {
				// This can happen if a file was deleted since the dir was listed. Unlikely but hey.
				continue
			} else if err != nil {
				c.err = err
				return
			}

			row := []FlowValueField{
				{Name: "name", Value: StringValue(entry.Name())},
				{Name: "type", Value: StringValue(util.Tern(entry.IsDir(), "dir", "file"))},
				{Name: "size", Value: Int64Value(info.Size(), FSUnitBytes)},
				{Name: "modified", Value: TimestampValue(info.ModTime())},
			}
			c.rows = append(c.rows, row)
		}
	}()

	return done
}

func (c *ListFilesAction) Result() NodeActionResult {
	return NodeActionResult{
		Err: c.err,
		Outputs: []FlowValue{{
			Type:       &FlowType{Kind: FSKindTable, ContainedType: FSFile},
			TableValue: c.rows,
		}},
	}
}

func StringValue(str string) *FlowValue {
	return &FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(str)}
}

func Int64Value(v int64, unit FlowUnit) *FlowValue {
	return &FlowValue{Type: &FlowType{Kind: FSKindInt64, Unit: unit}, Int64Value: v}
}

func TimestampValue(t time.Time) *FlowValue {
	return &FlowValue{Type: FSTimestamp, Int64Value: t.Unix()}
}
