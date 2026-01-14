package nodes

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type TrimSpacesAction struct{}

func NewTrimSpacesNode() *core.Node {
	return &core.Node{
		Name: "Trim Spaces",

		InputPorts: []core.NodePort{{
			Name: "Text",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Trimmed",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},

		Action: &TrimSpacesAction{},
	}
}

var _ core.NodeAction = &TrimSpacesAction{}

func (c *TrimSpacesAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	wire, hasWire := n.GetInputWire(0)
	if hasWire && core.Typecheck(wire.Type(), core.NewListType(core.FlowType{Kind: core.FSKindBytes})) == nil {
		n.InputPorts[0] = core.NodePort{
			Name: "Text items",
			Type: core.NewListType(core.FlowType{Kind: core.FSKindBytes}),
		}
		n.OutputPorts[0] = core.NodePort{
			Name: "Trimmed",
			Type: core.NewListType(core.FlowType{Kind: core.FSKindBytes}),
		}
	} else {
		n.InputPorts[0] = core.NodePort{
			Name: "Text",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}
		n.OutputPorts[0] = core.NodePort{
			Name: "Trimmed",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}
	}

	if !hasWire {
		n.Valid = false
	}
}

func (l *TrimSpacesAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("NodeContent", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:         core.GROWH,
			ChildAlignment: core.YCENTER,
		},
	}, func() {
		core.UIInputPort(n, 0)
		core.UISpacer(clay.IDI("TrimSpacesSpacer", n.ID), core.GROWH)
		core.UIOutputPort(n, 0)
	})
}

func (l *TrimSpacesAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)

	go func() {
		var res core.NodeActionResult
		defer func() {
			if r := recover(); r != nil {
				res = core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
			close(done)
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		input, ok, err := n.GetInputValue(0)
		if !ok {
			res.Err = errors.New("an input node is required")
			return
		}
		if err != nil {
			res.Err = err
			return
		}

		if input.Type.Kind == core.FSKindBytes {
			res = core.NodeActionResult{
				Outputs: []core.FlowValue{core.NewBytesValue(bytes.TrimSpace(input.BytesValue))},
			}
		} else {
			res = core.NodeActionResult{
				Outputs: []core.FlowValue{core.NewListValue(
					core.FlowType{Kind: core.FSKindBytes},
					util.Map(input.ListValue, func(fv core.FlowValue) core.FlowValue {
						return core.NewBytesValue(bytes.TrimSpace(fv.BytesValue))
					}),
				)},
			}
		}
	}()

	return done
}

func (l *TrimSpacesAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return l.RunContext(context.Background(), n)
}

func (n *TrimSpacesAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}
