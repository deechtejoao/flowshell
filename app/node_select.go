package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type SelectColumnsAction struct {
	SelectedColumns []string
}

func NewSelectColumnsNode() *Node {
	return &Node{
		Name: "Select Columns",

		InputPorts: []NodePort{{
			Name: "Input",
			Type: FlowType{Kind: FSKindTable},
		}},
		OutputPorts: []NodePort{{
			Name: "Output",
			Type: FlowType{Kind: FSKindTable},
		}},

		Action: &SelectColumnsAction{},
	}
}

var _ NodeAction = &SelectColumnsAction{}

func (c *SelectColumnsAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	
	input, wired := n.GetInputWire(0)
	if !wired {
		n.Valid = false
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable} // Generic table
		return
	}

	if input.Type().Kind != FSKindTable {
		n.Valid = false // Should be enforced by port type, but good to check
		return
	}

	// Calculate output type
	var newFields []FlowField
	if input.Type().ContainedType != nil {
		for _, field := range input.Type().ContainedType.Fields {
			if slices.Contains(c.SelectedColumns, field.Name) {
				newFields = append(newFields, field)
			}
		}
	}
	
	// If nothing selected, maybe default to all? Or none?
	// Let's say if list is empty, select all by default? 
	// Or maybe user explicitly wants empty table.
	// For now, if SelectedColumns is nil/empty, we might want to initialize it with all columns if this is the first run?
	// But UpdateAndValidate is called every frame. We shouldn't mutate state here if possible, or be careful.
	// Actually, initializing on first connect is a nice UX.
	
	// Let's rely on UI to let user select.
	
	n.OutputPorts[0].Type = FlowType{
		Kind: FSKindTable,
		ContainedType: &FlowType{
			Kind:   FSKindRecord,
			Fields: newFields,
		},
	}
}

func (c *SelectColumnsAction) UI(n *Node) {
	input, wired := n.GetInputWire(0)
	
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

		if wired && input.Type().Kind == FSKindTable {
			clay.CLAY_AUTO_ID(clay.EL{
				Layout: clay.LAY{
					LayoutDirection: clay.TopToBottom,
					ChildGap:        S1,
				},
			}, func() {
				clay.TEXT("Columns:", clay.TextElementConfig{TextColor: LightGray, FontSize: F1})
				
				for _, field := range input.Type().ContainedType.Fields {
					isSelected := slices.Contains(c.SelectedColumns, field.Name)
					
					UIButton(clay.AUTO_ID, UIButtonConfig{
						El: clay.EL{
							Layout: clay.LAY{ChildGap: S2, ChildAlignment: YCENTER},
						},
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							if isSelected {
								c.SelectedColumns = slices.DeleteFunc(c.SelectedColumns, func(s string) bool { return s == field.Name })
							} else {
								c.SelectedColumns = append(c.SelectedColumns, field.Name)
							}
							n.ClearResult()
						},
					}, func() {
						// Checkbox
						clay.CLAY_AUTO_ID(clay.EL{
							Layout:          clay.LAY{Sizing: WH(16, 16)},
							Border:          clay.B{Width: BA, Color: White},
							BackgroundColor: util.Tern(isSelected, Blue, clay.Color{}),
						})
						clay.TEXT(field.Name, clay.TextElementConfig{TextColor: White})
					})
				}
			})
		} else if !wired {
			clay.TEXT("Connect a table", clay.TextElementConfig{TextColor: Gray})
		}
	})
}

func (c *SelectColumnsAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				select {
				case done <- NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}:
				case <-ctx.Done():
				}
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

		// Map selected columns to indices
		var colIndices []int
		var newFields []FlowField
		
		if input.Type.ContainedType != nil {
			for i, field := range input.Type.ContainedType.Fields {
				if slices.Contains(c.SelectedColumns, field.Name) {
					colIndices = append(colIndices, i)
					newFields = append(newFields, field)
				}
			}
		}

		var newRows [][]FlowValueField
		for _, row := range input.TableValue {
			// Check context
			select {
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
				return
			default:
			}

			var newRow []FlowValueField
			for _, idx := range colIndices {
				newRow = append(newRow, row[idx])
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

func (c *SelectColumnsAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *SelectColumnsAction) Serialize(s *Serializer) bool {
	n := len(c.SelectedColumns)
	SInt(s, &n)
	if !s.Encode {
		c.SelectedColumns = make([]string, n)
	}
	for i := 0; i < n; i++ {
		SStr(s, &c.SelectedColumns[i])
	}
	return s.Ok()
}
