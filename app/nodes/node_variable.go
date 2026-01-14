package nodes

import (
	"context"
	"fmt"
	"os"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type GetVariableAction struct {
	VariableName string
}

func NewGetVariableNode() *core.Node {
	return &core.Node{
		Name: "Get Variable",
		OutputPorts: []core.NodePort{
			{Name: "Value", Type: core.FlowType{Kind: core.FSKindBytes}},
		},
		Action: &GetVariableAction{},
	}
}

var _ core.NodeAction = &GetVariableAction{}

func (a *GetVariableAction) UpdateAndValidate(n *core.Node) {
	if a.VariableName == "" {
		n.Valid = false
	} else {
		n.Valid = true
	}
}

func (a *GetVariableAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("GetVariable", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER, ChildGap: core.S2},
		}, func() {
			clay.TEXT("Name:", clay.TextElementConfig{TextColor: core.White})
			core.UITextBox(clay.IDI("VarName", n.ID), &a.VariableName, core.UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
			})
			core.UIOutputPort(n, 0)
		})
	})
}

func (a *GetVariableAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GetVariableAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	go func() {
		defer close(done)

		// Check environment variables first (higher priority? or lower? Let's say core.Graph Variables override Env?)
		// Actually, let's say Env overrides core.Graph so you can inject secrets at runtime.
		// Wait, "Variables / Secrets" usually implies managing them within the app.
		// Let's check:
		// 1. System Env Vars
		// 2. core.Graph Variables

		val, found := os.LookupEnv(a.VariableName)
		if !found {
			if n.Graph != nil {
				n.Graph.VarMutex.RLock()
				v, ok := n.Graph.Variables[a.VariableName]
				n.Graph.VarMutex.RUnlock()
				if ok {
					val = v
					found = true
				}
			}
		}

		if !found {
			done <- core.NodeActionResult{Err: fmt.Errorf("variable '%s' not found in environment or graph", a.VariableName)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{core.NewStringValue(val)},
		}
	}()
	return done
}

func (a *GetVariableAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &a.VariableName)
	return s.Ok()
}