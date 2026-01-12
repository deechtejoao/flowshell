package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type TransposeAction struct{}

func (c *TransposeAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

func NewTransposeNode() *Node {
	return &Node{
		Name: "Transpose",

		InputPorts: []NodePort{{
			Name: "Input",
			Type: FlowType{Kind: FSKindTable},
		}},
		OutputPorts: []NodePort{{
			Name: "Output",
			Type: FlowType{Kind: FSKindTable},
		}},

		Action: &TransposeAction{},
	}
}

var _ NodeAction = &TransposeAction{}

func (c *TransposeAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if !wired {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable}
		return
	}
	if input.Type().Kind != FSKindTable {
		n.Valid = false
		return
	}

	// We can't know the output schema without the data (number of rows determines number of columns)
	// So we'll just set it to Any or generic Table for now.
	// Actually, UpdateAndValidate is for static typing. Transpose changes schema based on data size (rows -> cols).
	// So we can't statically determine the fields of the output table if the number of rows is dynamic.
	// We'll mark the output as a Table with unknown structure (nil ContainedType? or empty Fields?).
	// FlowType definition:
	// type FlowType struct { Kind FSKind; ContainedType *FlowType; ... }
	// If Kind is Table, ContainedType should be Record describing the row.
	// If we don't know the fields, maybe we leave Fields empty?

	// For now, let's just say it returns a Table.
	n.OutputPorts[0].Type = FlowType{Kind: FSKindTable, ContainedType: &FlowType{Kind: FSKindRecord}}
}

func (c *TransposeAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{
				Sizing:         GROWH,
				ChildAlignment: YCENTER,
			},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})
	})
}

func (c *TransposeAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				done <- NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
		}()

		select {
		case <-ctx.Done():
			done <- NodeActionResult{Err: ctx.Err()}
			return
		default:
		}

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		if input.Type.Kind != FSKindTable {
			done <- NodeActionResult{Err: fmt.Errorf("input must be a table")}
			return
		}

		// Old Table:
		// Cols: C1, C2, C3
		// Rows:
		// R1: v11, v12, v13
		// R2: v21, v22, v23

		// New Table:
		// Cols: Row 0, Row 1 (based on R1, R2)
		// Rows:
		// NR1 (was C1): v11, v21
		// NR2 (was C2): v12, v22
		// NR3 (was C3): v13, v23

		numOldRows := len(input.TableValue)
		numOldCols := len(input.Type.ContainedType.Fields)

		if numOldRows == 0 {
			// Empty table
			done <- NodeActionResult{
				Outputs: []FlowValue{input},
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

		var newFields []FlowField
		for i := 0; i < numOldRows; i++ {
			// Check context
			select {
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
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
				kind = FSKindAny
			}

			newFields = append(newFields, FlowField{
				Name: "Row " + strconv.Itoa(i),
				Type: &FlowType{Kind: kind},
			})
		}

		// Build new rows
		var newRows [][]FlowValueField
		for c := 0; c < numOldCols; c++ {
			// Check context
			select {
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
				return
			default:
			}

			var newRow []FlowValueField
			for r := 0; r < numOldRows; r++ {
				val := input.TableValue[r][c].Value
				// If new column is Any but val is specific, that's fine, Value stores its own type.
				// But if we want to enforce schema, we might need to wrap/cast?
				// FlowValue is self-describing, so it's fine.

				newRow = append(newRow, FlowValueField{
					Name:  newFields[r].Name,
					Value: val,
				})
			}
			newRows = append(newRows, newRow)
		}

		done <- NodeActionResult{
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

func (c *TransposeAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}
