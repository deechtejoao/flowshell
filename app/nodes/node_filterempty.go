package nodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type FilterEmptyAction struct {
	Column string

	// UI State
	dropdown core.UIDropdown
}

func NewFilterEmptyNode() *core.Node {
	return &core.Node{
		Name: "Filter Empty",

		InputPorts: []core.NodePort{{
			Name: "Input",
			Type: core.FlowType{Kind: core.FSKindTable},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Filtered",
			Type: core.FlowType{Kind: core.FSKindTable},
		}},

		Action: &FilterEmptyAction{},
	}
}

var _ core.NodeAction = &FilterEmptyAction{}

func (c *FilterEmptyAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if wired {
		n.OutputPorts[0].Type = input.Type()
		if input.Type().Kind != core.FSKindTable {
			n.Valid = false
		}
	} else {
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable}
		n.Valid = false
	}
}

func (c *FilterEmptyAction) UI(n *core.Node) {
	input, wired := n.GetInputWire(0)

	clay.CLAY(clay.IDI("FilterEmptyContainer", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("FilterEmptyRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("FilterEmptySpacer", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		if wired && input.Type().Kind == core.FSKindTable {
			var options []core.UIDropdownOption
			for _, field := range input.Type().ContainedType.Fields {
				options = append(options, core.UIDropdownOption{
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

			c.dropdown.Do(clay.IDI("FilterColumn", n.ID), core.UIDropdownConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
				OnChange: func(_, after any) {
					c.Column = after.(string)
				},
			})
		}
	})
}

func (c *FilterEmptyAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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
		if !ok || err != nil {
			res.Err = err
			return
		}

		if input.Type.Kind != core.FSKindTable {
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

		var newRows [][]core.FlowValueField
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
			case core.FSKindBytes:
				if len(val.BytesValue) == 0 {
					keep = false
				}
			case core.FSKindInt64:
				if val.Int64Value == 0 {
					keep = false
				}
			case core.FSKindFloat64:
				if val.Float64Value == 0 {
					keep = false
				}
			case core.FSKindList:
				if len(val.ListValue) == 0 {
					keep = false
				}
			case core.FSKindRecord:
				if len(val.RecordValue) == 0 {
					keep = false
				}
			case core.FSKindTable:
				if len(val.TableValue) == 0 {
					keep = false
				}
			}

			if keep {
				newRows = append(newRows, row)
			}
		}

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:       input.Type,
				TableValue: newRows,
			}},
		}
	}()
	return done
}

func (c *FilterEmptyAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *FilterEmptyAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.Column)
	return s.Ok()
}