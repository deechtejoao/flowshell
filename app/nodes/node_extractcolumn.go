package nodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type ExtractColumnAction struct {
	Column string

	// UI State
	dropdown core.UIDropdown
}

func NewExtractColumnNode() *core.Node {
	return &core.Node{
		Name: "Extract Column",

		InputPorts: []core.NodePort{{
			Name: "Input",
			Type: core.FlowType{Kind: core.FSKindTable},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Column",
			Type: core.FlowType{Kind: core.FSKindList}, // Will be refined in UpdateAndValidate
		}},

		Action: &ExtractColumnAction{},
	}
}

var _ core.NodeAction = &ExtractColumnAction{}

func (c *ExtractColumnAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)

	if !wired {
		n.Valid = false
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindList, ContainedType: &core.FlowType{Kind: core.FSKindAny}}
		return
	}

	if input.Type().Kind != core.FSKindTable {
		n.Valid = false
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindList, ContainedType: &core.FlowType{Kind: core.FSKindAny}}
		return
	}

	// Find the column type
	found := false
	for _, field := range input.Type().ContainedType.Fields {
		if field.Name == c.Column {
			n.OutputPorts[0].Type = core.FlowType{
				Kind:          core.FSKindList,
				ContainedType: field.Type,
			}
			found = true
			break
		}
	}

	if !found {
		// If column not found (or not selected yet), default to Any
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindList, ContainedType: &core.FlowType{Kind: core.FSKindAny}}
	}
}

func (c *ExtractColumnAction) UI(n *core.Node) {
	input, wired := n.GetInputWire(0)

	clay.CLAY(clay.IDI("ExtractColumnUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("ExtractColumnRow", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("ExtractColumnSpacer", n.ID), core.GROWH)
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

			c.dropdown.Do(clay.IDI("ExtractColumn", n.ID), core.UIDropdownConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
				OnChange: func(_, after any) {
					c.Column = after.(string)
				},
			})
		}
	})
}

func (c *ExtractColumnAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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
		var colType *core.FlowType
		for i, field := range input.Type.ContainedType.Fields {
			if field.Name == c.Column {
				colIdx = i
				colType = field.Type
				break
			}
		}

		if colIdx == -1 {
			res.Err = fmt.Errorf("column %q not found", c.Column)
			return
		}

		var list []core.FlowValue
		for _, row := range input.TableValue {
			// Check context
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				return
			default:
			}

			list = append(list, row[colIdx].Value)
		}

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:      &core.FlowType{Kind: core.FSKindList, ContainedType: colType},
				ListValue: list,
			}},
		}
	}()
	return done
}

func (c *ExtractColumnAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *ExtractColumnAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.Column)
	return s.Ok()
}