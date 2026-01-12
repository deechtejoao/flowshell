package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type AddColumnAction struct {
	NewColumnName string
}

func NewAddColumnNode() *Node {
	return &Node{
		ID:   NewNodeID(),
		Name: "Add Column",

		InputPorts: []NodePort{
			{Name: "Table", Type: FlowType{Kind: FSKindTable}},
			{Name: "Values", Type: FlowType{Kind: FSKindList}},
		},
		OutputPorts: []NodePort{
			{Name: "Table", Type: FlowType{Kind: FSKindTable}},
		},

		Action: &AddColumnAction{NewColumnName: "New Column"},
	}
}

var _ NodeAction = &AddColumnAction{}

func (c *AddColumnAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	tableWire, tableWired := n.GetInputWire(0)
	valuesWire, valuesWired := n.GetInputWire(1)

	if !tableWired || !valuesWired {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable, ContainedType: &FlowType{Kind: FSKindAny}} // Unknown
		return
	}

	if tableWire.Type().Kind != FSKindTable || valuesWire.Type().Kind != FSKindList {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable, ContainedType: &FlowType{Kind: FSKindAny}}
		return
	}

	// Construct new table type
	originalFields := tableWire.Type().ContainedType.Fields
	newField := FlowField{
		Name: c.NewColumnName,
		Type: valuesWire.Type().ContainedType,
	}

	newFields := make([]FlowField, len(originalFields)+1)
	copy(newFields, originalFields)
	newFields[len(originalFields)] = newField

	n.OutputPorts[0].Type = FlowType{
		Kind: FSKindTable,
		ContainedType: &FlowType{
			Kind:   FSKindRecord,
			Fields: newFields,
		},
	}
}

func (c *AddColumnAction) UI(n *Node) {
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

		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 1)
			UISpacer(clay.AUTO_ID, GROWH)
		})

		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("NewColumnName", n.ID), &c.NewColumnName, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})
	})
}

func (c *AddColumnAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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

		if tableInput.Type.Kind != FSKindTable {
			res.Err = errors.New("input 1 must be a table")
			return
		}
		if valuesInput.Type.Kind != FSKindList {
			res.Err = errors.New("input 2 must be a list")
			return
		}

		if len(tableInput.TableValue) != len(valuesInput.ListValue) {
			res.Err = fmt.Errorf("row count mismatch: table has %d, values list has %d", len(tableInput.TableValue), len(valuesInput.ListValue))
			return
		}

		var newRows [][]FlowValueField
		for i, row := range tableInput.TableValue {
			select {
			case <-ctx.Done():
				res.Err = ctx.Err()
				return
			default:
			}

			newRow := make([]FlowValueField, len(row)+1)
			copy(newRow, row)
			newRow[len(row)] = FlowValueField{
				Name:  c.NewColumnName,
				Value: valuesInput.ListValue[i],
			}
			newRows = append(newRows, newRow)
		}

		// Output type should match what we calculated in UpdateAndValidate, but let's reconstruct it to be safe
		originalFields := tableInput.Type.ContainedType.Fields
		newField := FlowField{
			Name: c.NewColumnName,
			Type: valuesInput.Type.ContainedType,
		}
		newFields := make([]FlowField, len(originalFields)+1)
		copy(newFields, originalFields)
		newFields[len(originalFields)] = newField

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type: &FlowType{
					Kind: FSKindTable,
					ContainedType: &FlowType{
						Kind:   FSKindRecord,
						Fields: newFields,
					},
				},
				TableValue: newRows,
			}},
		}
	}()
	return done
}

func (c *AddColumnAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *AddColumnAction) Serialize(s *Serializer) bool {
	SStr(s, &c.NewColumnName)
	return s.Ok()
}
