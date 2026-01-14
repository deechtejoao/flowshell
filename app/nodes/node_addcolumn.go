package nodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type AddColumnAction struct {
	NewColumnName string
}

func NewAddColumnNode() *core.Node {
	return &core.Node{
		Name: "Add Column",

		InputPorts: []core.NodePort{
			{Name: "Table", Type: core.FlowType{Kind: core.FSKindTable}},
			{Name: "Values", Type: core.FlowType{Kind: core.FSKindList}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Table", Type: core.FlowType{Kind: core.FSKindTable}},
		},

		Action: &AddColumnAction{NewColumnName: "New Column"},
	}
}

var _ core.NodeAction = &AddColumnAction{}

func (c *AddColumnAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	tableWire, tableWired := n.GetInputWire(0)
	valuesWire, valuesWired := n.GetInputWire(1)

	if !tableWired || !valuesWired {
		n.Valid = false
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable, ContainedType: &core.FlowType{Kind: core.FSKindAny}} // Unknown
		return
	}

	if tableWire.Type().Kind != core.FSKindTable || valuesWire.Type().Kind != core.FSKindList {
		n.Valid = false
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable, ContainedType: &core.FlowType{Kind: core.FSKindAny}}
		return
	}

	// Construct new table type
	originalFields := tableWire.Type().ContainedType.Fields
	newField := core.FlowField{
		Name: c.NewColumnName,
		Type: valuesWire.Type().ContainedType,
	}

	newFields := make([]core.FlowField, len(originalFields)+1)
	copy(newFields, originalFields)
	newFields[len(originalFields)] = newField

	n.OutputPorts[0].Type = core.FlowType{
		Kind: core.FSKindTable,
		ContainedType: &core.FlowType{
			Kind:   core.FSKindRecord,
			Fields: newFields,
		},
	}
}

func (c *AddColumnAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("AddColumnUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("AddColumnRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("AddColumnSpacer1", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("AddColumnRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 1)
			core.UISpacer(clay.IDI("AddColumnSpacer2", n.ID), core.GROWH)
		})

		clay.CLAY(clay.IDI("AddColumnRow3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH},
		}, func() {
			core.UITextBox(clay.IDI("NewColumnName", n.ID), &c.NewColumnName, core.UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
			})
		})
	})
}

func (c *AddColumnAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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

		tableInput, ok1, err1 := n.GetInputValue(0)
		valuesInput, ok2, err2 := n.GetInputValue(1)

		if !ok1 || !ok2 {
			res.Err = errors.New("missing inputs")
			return
		}
		if err1 != nil {
			res.Err = err1
			return
		}
		if err2 != nil {
			res.Err = err2
			return
		}

		if tableInput.Type.Kind != core.FSKindTable {
			res.Err = errors.New("input 1 must be a table")
			return
		}
		if valuesInput.Type.Kind != core.FSKindList {
			res.Err = errors.New("input 2 must be a list")
			return
		}

		if len(tableInput.TableValue) != len(valuesInput.ListValue) {
			res.Err = fmt.Errorf("row count mismatch: table has %d, values list has %d", len(tableInput.TableValue), len(valuesInput.ListValue))
			return
		}

		var newRows [][]core.FlowValueField
		for i, row := range tableInput.TableValue {
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				return
			default:
			}

			newRow := make([]core.FlowValueField, len(row)+1)
			copy(newRow, row)
			newRow[len(row)] = core.FlowValueField{
				Name:  c.NewColumnName,
				Value: valuesInput.ListValue[i],
			}
			newRows = append(newRows, newRow)
		}

		// Output type should match what we calculated in UpdateAndValidate, but let's reconstruct it to be safe
		originalFields := tableInput.Type.ContainedType.Fields
		newField := core.FlowField{
			Name: c.NewColumnName,
			Type: valuesInput.Type.ContainedType,
		}
		newFields := make([]core.FlowField, len(originalFields)+1)
		copy(newFields, originalFields)
		newFields[len(originalFields)] = newField

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type: &core.FlowType{
					Kind: core.FSKindTable,
					ContainedType: &core.FlowType{
						Kind:   core.FSKindRecord,
						Fields: newFields,
					},
				},
				TableValue: newRows,
			}},
		}
	}()
	return done
}

func (c *AddColumnAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *AddColumnAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.NewColumnName)
	return s.Ok()
}