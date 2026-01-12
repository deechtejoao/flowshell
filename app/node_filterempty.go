package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type FilterEmptyAction struct {
	Column string

	// UI State
	dropdown UIDropdown
}

func NewFilterEmptyNode() *Node {
	return &Node{
		Name: "Filter Empty",

		InputPorts: []NodePort{{
			Name: "Input",
			Type: FlowType{Kind: FSKindTable},
		}},
		OutputPorts: []NodePort{{
			Name: "Filtered",
			Type: FlowType{Kind: FSKindTable},
		}},

		Action: &FilterEmptyAction{},
	}
}

var _ NodeAction = &FilterEmptyAction{}

func (c *FilterEmptyAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if wired {
		n.OutputPorts[0].Type = input.Type()
		if input.Type().Kind != FSKindTable {
			n.Valid = false
		}
	} else {
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable}
		n.Valid = false
	}
}

func (c *FilterEmptyAction) UI(n *Node) {
	input, wired := n.GetInputWire(0)

	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})

		if wired && input.Type().Kind == FSKindTable {
			var options []UIDropdownOption
			for _, field := range input.Type().ContainedType.Fields {
				options = append(options, UIDropdownOption{
					Name:  field.Name,
					Value: field.Name,
				})
			}

			// Update options if they changed
			// Simple check: different length or different first/last?
			// For now, let's just replace them.
			// TODO: Preserve selection better if columns shift but same names exist?
			c.dropdown.Options = options

			// Sync selection
			if c.Column == "" {
				if len(options) > 0 {
					c.Column = options[0].Value.(string)
					c.dropdown.Selected = 0
				}
			} else {
				if !c.dropdown.SelectByValue(c.Column) {
					// Selection no longer valid, reset to first option
					if len(options) > 0 {
						c.Column = options[0].Value.(string)
						c.dropdown.Selected = 0
					} else {
						c.Column = ""
					}
				}
			}

			c.dropdown.Do(clay.IDI("FilterColumn", n.ID), UIDropdownConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
				OnChange: func(_, after any) {
					c.Column = after.(string)
				},
			})
		}
	})
}

func (c *FilterEmptyAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			res.Err = err
			return
		}

		if input.Type.Kind != FSKindTable {
			res.Err = errors.New("input must be a table")
			return
		}

		colIdx := -1
		for i, field := range input.Type.ContainedType.Fields {
			if field.Name == c.Column {
				colIdx = i
				break
			}
		}

		if colIdx == -1 {
			// If column not found, we can't filter.
			// Return error.
			res.Err = fmt.Errorf("column %q not found", c.Column)
			return
		}

		var newRows [][]FlowValueField
		for _, row := range input.TableValue {
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				return
			default:
			}

			val := row[colIdx].Value
			keep := true

			switch val.Type.Kind {
			case FSKindBytes:
				if len(val.BytesValue) == 0 {
					keep = false
				}
			case FSKindInt64:
				if val.Int64Value == 0 {
					keep = false
				}
			case FSKindFloat64:
				if val.Float64Value == 0 {
					keep = false
				}
			case FSKindList:
				if len(val.ListValue) == 0 {
					keep = false
				}
			case FSKindRecord:
				if len(val.RecordValue) == 0 {
					keep = false
				}
			case FSKindTable:
				if len(val.TableValue) == 0 {
					keep = false
				}
			}

			if keep {
				newRows = append(newRows, row)
			}
		}

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:       input.Type,
				TableValue: newRows,
			}},
		}
	}()
	return done
}

func (c *FilterEmptyAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *FilterEmptyAction) Serialize(s *Serializer) bool {
	SStr(s, &c.Column)
	return s.Ok()
}
