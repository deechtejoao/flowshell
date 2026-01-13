package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ConcatTablesAction struct{}

func NewConcatTablesNode() *Node {
	return &Node{
		Name: "Concatenate Tables",

		InputPorts: []NodePort{{
			Name: "Table 1",
			Type: NewAnyTableType(),
		}},
		OutputPorts: []NodePort{{
			Name: "Table",
			Type: NewAnyTableType(),
		}},

		Action: &ConcatTablesAction{},
	}
}

var _ NodeAction = &ConcatTablesAction{}

func (a *ConcatTablesAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	allPortsWired := true
	for i := range n.InputPorts {
		wire, hasWire := n.GetInputWire(i)
		if !hasWire {
			allPortsWired = false
			n.OutputPorts[0].Type = NewAnyTableType()
		} else if i == 0 {
			n.OutputPorts[0].Type = wire.Type()
		}
	}

	if !allPortsWired {
		n.Valid = false
	}
}

func (a *ConcatTablesAction) UI(n *Node) {
	clay.CLAY(clay.IDI("ConcatTablesUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("ConcatTablesRow1", n.ID), clay.EL{
			Layout: clay.LAY{
				Sizing:         GROWH,
				ChildAlignment: YCENTER,
			},
		}, func() {
			buttonStyle := clay.EL{
				Layout: clay.LAY{
					Sizing:         WH(24, 24),
					ChildAlignment: ALLCENTER,
				},
				Border: clay.B{Width: BA, Color: Gray},
			}
			buttonTextConfig := clay.T{FontID: InterSemibold, FontSize: F2, TextColor: White}

			UIButton(clay.IDI("ConcatMinus", n.ID), UIButtonConfig{ // -
				El: buttonStyle,
				OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
					if len(n.InputPorts) > 1 {
						if n.Graph != nil {
							n.Graph.Wires = slices.DeleteFunc(n.Graph.Wires, func(w *Wire) bool {
								return w.EndNode == n && w.EndPort >= len(n.InputPorts)-1
							})
						}
						n.InputPorts = n.InputPorts[:len(n.InputPorts)-1]
					}
				},
			}, func() {
				clay.TEXT("-", buttonTextConfig)
			})
			UISpacer(clay.IDI("ConcatMinusSpacer", n.ID), W1)
			UIButton(clay.IDI("ConcatPlus", n.ID), UIButtonConfig{ // +
				El: buttonStyle,
				OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
					n.InputPorts = append(n.InputPorts, NodePort{
						Name: fmt.Sprintf("Table %d", len(n.InputPorts)+1),
						Type: NewAnyTableType(),
					})
				},
			}, func() {
				clay.TEXT("+", buttonTextConfig)
			})
			UISpacer(clay.IDI("ConcatPlusSpacer", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		for i := range n.InputPorts {
			UIInputPort(n, i)
		}
	})
}

func (a *ConcatTablesAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		var res NodeActionResult
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
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

		var finalTableRows [][]FlowValueField
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

			if err := Typecheck(*input.Type, *expectedType); err != nil {
				res.Err = fmt.Errorf("all tables should have the same type: expected %s but got %s", expectedType, input.Type)
				return
			}
			finalTableRows = append(finalTableRows, input.TableValue...)
		}

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:       expectedType,
				TableValue: finalTableRows,
			}},
		}
	}()
	return done
}

func (a *ConcatTablesAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (n *ConcatTablesAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
