package nodes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type TransposeAction struct{}

func (c *TransposeAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

func NewTransposeNode() *core.Node {
	return &core.Node{
		Name: "Transpose",

		InputPorts: []core.NodePort{{
			Name: "Input",
			Type: core.FlowType{Kind: core.FSKindTable},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Output",
			Type: core.FlowType{Kind: core.FSKindTable},
		}},

		Action: &TransposeAction{},
	}
}

var _ core.NodeAction = &TransposeAction{}

func (c *TransposeAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if !wired {
		n.Valid = false
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable}
		return
	}
	if input.Type().Kind != core.FSKindTable {
		n.Valid = false
		return
	}

	// We can't know the output schema without the data (number of rows determines number of columns)
	// So we'll just set it to Any or generic Table for now.
	// Actually, UpdateAndValidate is for static typing. Transpose changes schema based on data size (rows -> cols).
	// So we can't statically determine the fields of the output table if the number of rows is dynamic.
	// We'll mark the output as a Table with unknown structure (nil ContainedType? or empty Fields?).
	// core.FlowType definition:
	// type core.FlowType struct { Kind FSKind; ContainedType *core.FlowType; ... }
	// If Kind is Table, ContainedType should be Record describing the row.
	// If we don't know the fields, maybe we leave Fields empty?

	// For now, let's just say it returns a Table.
	n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable, ContainedType: &core.FlowType{Kind: core.FSKindRecord}}
}

func (c *TransposeAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("TransposeUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("TransposePorts", n.ID), clay.EL{
			Layout: clay.LAY{
				Sizing:         core.GROWH,
				ChildAlignment: core.YCENTER,
			},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("TransposeSpacer", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})
	})
}

func (c *TransposeAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}

		if input.Type.Kind != core.FSKindTable {
			done <- core.NodeActionResult{Err: fmt.Errorf("input must be a table")}
			return
		}

		// Old Table:
		// Cols: C1, C2, C3
		// Rows:
		// core.R1: v11, v12, v13
		// core.R2: v21, v22, v23

		// New Table:
		// Cols: Row 0, Row 1 (based on core.R1, core.R2)
		// Rows:
		// NR1 (was C1): v11, v21
		// NR2 (was C2): v12, v22
		// NR3 (was C3): v13, v23

		numOldRows := len(input.TableValue)
		numOldCols := len(input.Type.ContainedType.Fields)

		if numOldRows == 0 {
			// Empty table
			done <- core.NodeActionResult{
				Outputs: []core.FlowValue{input},
			}
			return
		}

		// Determine new column types
		// New Col j corresponds to Old Row j.
		// Old Row j contains values from Old Cols 0..N.
		// If Old Cols have different types, New Col j will have mixed types -> Any.

		// Wait.
		// New Col j = [v_j_0, v_j_1, ... v_j_oldCols] -- NO.
		// New Col j corresponds to Old Row j.
		// The values in New Col j are: v_j_0 (from old col 0), v_j_1 (from old col 1), etc.
		// So New Col j contains values from specific row j across all old columns.
		// So if old columns have different types, New Col j mixes them.

		var newFields []core.FlowField
		for i := 0; i < numOldRows; i++ {
			// Check context
			select {
			case <-ctx.Done():
				done <- core.NodeActionResult{Err: ctx.Err()}
				return
			default:
			}

			// Check types in this row
			firstType := input.TableValue[i][0].Value.Type
			mixed := false
			for j := 1; j < numOldCols; j++ {
				if input.TableValue[i][j].Value.Type.Kind != firstType.Kind {
					mixed = true
					break
				}
			}

			kind := firstType.Kind
			if mixed {
				kind = core.FSKindAny
			}

			newFields = append(newFields, core.FlowField{
				Name: "Row " + strconv.Itoa(i),
				Type: &core.FlowType{Kind: kind},
			})
		}

		// Build new rows
		var newRows [][]core.FlowValueField
		for c := 0; c < numOldCols; c++ {
			// Check context
			select {
			case <-ctx.Done():
				done <- core.NodeActionResult{Err: ctx.Err()}
				return
			default:
			}

			var newRow []core.FlowValueField
			for r := 0; r < numOldRows; r++ {
				val := input.TableValue[r][c].Value
				// If new column is Any but val is specific, that's fine, Value stores its own type.
				// But if we want to enforce schema, we might need to wrap/cast?
				// core.FlowValue is self-describing, so it's fine.

				newRow = append(newRow, core.FlowValueField{
					Name:  newFields[r].Name,
					Value: val,
				})
			}
			newRows = append(newRows, newRow)
		}

		done <- core.NodeActionResult{
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

func (c *TransposeAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}