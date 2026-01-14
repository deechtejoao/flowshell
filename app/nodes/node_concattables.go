package nodes

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type ConcatTablesAction struct{}

func NewConcatTablesNode() *core.Node {
	return &core.Node{
		Name: "Concatenate Tables",

		InputPorts: []core.NodePort{{
			Name: "Table 1",
			Type: core.NewAnyTableType(),
		}},
		OutputPorts: []core.NodePort{{
			Name: "Table",
			Type: core.NewAnyTableType(),
		}},

		Action: &ConcatTablesAction{},
	}
}

var _ core.NodeAction = &ConcatTablesAction{}

func (a *ConcatTablesAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	allPortsWired := true
	for i := range n.InputPorts {
		wire, hasWire := n.GetInputWire(i)
		if !hasWire {
			allPortsWired = false
			n.OutputPorts[0].Type = core.NewAnyTableType()
		} else if i == 0 {
			n.OutputPorts[0].Type = wire.Type()
		}
	}

	if !allPortsWired {
		n.Valid = false
	}
}

func (a *ConcatTablesAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("ConcatTablesUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("ConcatTablesRow1", n.ID), clay.EL{
			Layout: clay.LAY{
				Sizing:         core.GROWH,
				ChildAlignment: core.YCENTER,
			},
		}, func() {
			buttonStyle := clay.EL{
				Layout: clay.LAY{
					Sizing:         core.WH(24, 24),
					ChildAlignment: core.ALLCENTER,
				},
				Border: clay.B{Width: core.BA, Color: core.Gray},
			}
			buttonTextConfig := clay.T{FontID: core.InterSemibold, FontSize: core.F2, TextColor: core.White}

			core.UIButton(clay.IDI("ConcatMinus", n.ID), core.UIButtonConfig{ // -
				El: buttonStyle,
				OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
					if len(n.InputPorts) > 1 {
						if n.Graph != nil {
							n.Graph.Wires = slices.DeleteFunc(n.Graph.Wires, func(w *core.Wire) bool {
								return w.EndNode == n && w.EndPort >= len(n.InputPorts)-1
							})
						}
						n.InputPorts = n.InputPorts[:len(n.InputPorts)-1]
					}
				},
			}, func() {
				clay.TEXT("-", buttonTextConfig)
			})
			core.UISpacer(clay.IDI("ConcatMinusSpacer", n.ID), core.W1)
			core.UIButton(clay.IDI("ConcatPlus", n.ID), core.UIButtonConfig{ // +
				El: buttonStyle,
				OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
					n.InputPorts = append(n.InputPorts, core.NodePort{
						Name: fmt.Sprintf("Table %d", len(n.InputPorts)+1),
						Type: core.NewAnyTableType(),
					})
				},
			}, func() {
				clay.TEXT("+", buttonTextConfig)
			})
			core.UISpacer(clay.IDI("ConcatPlusSpacer", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		for i := range n.InputPorts {
			core.UIInputPort(n, i)
		}
	})
}

func (a *ConcatTablesAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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

		firstInput, ok, err := n.GetInputValue(0)
		if !ok {
			panic("first input was not wired (validation should have caught this)")
		}
		if err != nil {
			res.Err = err
			return
		}
		expectedType := firstInput.Type

		var finalTableRows [][]core.FlowValueField
		for i := range n.InputPorts {
			// Check context
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				return
			default:
			}

			input, ok, err := n.GetInputValue(i)
			if !ok {
				res.Err = errors.New("an input node is required")
				return
			}
			if err != nil {
				res.Err = err
				return
			}

			if err := core.Typecheck(*input.Type, *expectedType); err != nil {
				res.Err = fmt.Errorf("all tables should have the same type: expected %s but got %s", expectedType, input.Type)
				return
			}
			finalTableRows = append(finalTableRows, input.TableValue...)
		}

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:       expectedType,
				TableValue: finalTableRows,
			}},
		}
	}()
	return done
}

func (a *ConcatTablesAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (n *ConcatTablesAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}
