package nodes

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ValueAction struct {
	Value core.FlowValue
}

func NewValueNode(v core.FlowValue) *core.Node {
	return &core.Node{
		Name: "Value",

		OutputPorts: []core.NodePort{{
			Name: "Value",
			Type: *v.Type,
		}},

		Action: &ValueAction{Value: v},
	}
}

var _ core.NodeAction = &ValueAction{}

func (c *ValueAction) Serialize(s *core.Serializer) bool {
	core.SThing(s, &c.Value)
	return s.Ok()
}

func (c *ValueAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	// Ensure output port type matches value type (in case it changed or for safety)
	if n.OutputPorts[0].Type.Kind == core.FSKindAny {
		n.OutputPorts[0].Type = *c.Value.Type
	}
}

func (c *ValueAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("NodeContent", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:         core.GROWH,
			ChildAlignment: core.YCENTER,
		},
	}, func() {
		// Just show the value
		core.UIFlowValue(c.Value)
		core.UISpacer(clay.IDI("ValueSpacer", n.ID), core.GROWH)
		core.UIOutputPort(n, 0)
	})
}

func (c *ValueAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				done <- core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
		}()

		select {
		case <-ctx.Done():
			done <- core.NodeActionResult{Err: ctx.Err()}
			return
		default:
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{c.Value},
		}
	}()
	return done
}

func (c *ValueAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}
