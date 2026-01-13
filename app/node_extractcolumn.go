package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ExtractColumnAction struct {
	Column string

	// UI State
	dropdown UIDropdown
}

func NewExtractColumnNode() *Node {
	return &Node{
		Name: "Extract Column",

		InputPorts: []NodePort{{
			Name: "Input",
			Type: FlowType{Kind: FSKindTable},
		}},
		OutputPorts: []NodePort{{
			Name: "Column",
			Type: FlowType{Kind: FSKindList}, // Will be refined in UpdateAndValidate
		}},

		Action: &ExtractColumnAction{},
	}
}

var _ NodeAction = &ExtractColumnAction{}

func (c *ExtractColumnAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)

	if !wired {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindAny}}
		return
	}

	if input.Type().Kind != FSKindTable {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindAny}}
		return
	}

	// Find the column type
	found := false
	for _, field := range input.Type().ContainedType.Fields {
		if field.Name == c.Column {
			n.OutputPorts[0].Type = FlowType{
				Kind:          FSKindList,
				ContainedType: field.Type,
			}
			found = true
			break
		}
	}

	if !found {
		// If column not found (or not selected yet), default to Any
		n.OutputPorts[0].Type = FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindAny}}
	}
}

func (c *ExtractColumnAction) UI(n *Node) {
	input, wired := n.GetInputWire(0)

	clay.CLAY(clay.IDI("ExtractColumnUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("ExtractColumnRow", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("ExtractColumnSpacer", n.ID), GROWH)
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

			c.dropdown.Do(clay.IDI("ExtractColumn", n.ID), UIDropdownConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
				OnChange: func(_, after any) {
					c.Column = after.(string)
				},
			})
		}
	})
}

func (c *ExtractColumnAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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
		var colType *FlowType
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

		var list []FlowValue
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

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:      &FlowType{Kind: FSKindList, ContainedType: colType},
				ListValue: list,
			}},
		}
	}()
	return done
}

func (c *ExtractColumnAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *ExtractColumnAction) Serialize(s *Serializer) bool {
	SStr(s, &c.Column)
	return s.Ok()
}
