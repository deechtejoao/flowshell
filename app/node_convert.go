package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type ConvertAction struct {
	TargetKind     FlowTypeKind
	targetDropdown UIDropdown

	Column      string
	colDropdown UIDropdown

	IgnoreErrors bool
}

var convertOptions = []UIDropdownOption{
	{Name: "Bytes (String)", Value: FSKindBytes},
	{Name: "Int64", Value: FSKindInt64},
	{Name: "Float64", Value: FSKindFloat64},
}

func NewConvertNode() *Node {
	action := ConvertAction{
		TargetKind: FSKindBytes, // Default
		targetDropdown: UIDropdown{
			Options: convertOptions,
		},
		colDropdown: UIDropdown{
			Options: []UIDropdownOption{},
		},
	}
	action.targetDropdown.SelectByValue(FSKindBytes)

	return &Node{
		ID:   NewNodeID(),
		Name: "Convert Type",

		InputPorts: []NodePort{{
			Name: "Input",
			Type: FlowType{Kind: FSKindAny},
		}},
		OutputPorts: []NodePort{{
			Name: "Output",
			Type: FlowType{Kind: FSKindBytes},
		}},

		Action: &action,
	}
}

var _ NodeAction = &ConvertAction{}

func (c *ConvertAction) Serialize(s *Serializer) bool {
	SInt(s, (*int)(&c.TargetKind))
	SStr(s, &c.Column)
	SBool(s, &c.IgnoreErrors)
	return s.Ok()
}

func (c *ConvertAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	// Ensure target dropdown matches state
	if c.targetDropdown.Selected == -1 {
		c.targetDropdown.SelectByValue(c.TargetKind)
	} else {
		c.TargetKind = c.targetDropdown.GetSelectedOption().Value.(FlowTypeKind)
	}

	// Update input/output types and column dropdown
	if n.InputIsWired(0) {
		inputType := n.InputPorts[0].Type
		if inputType.Kind == FSKindTable && inputType.ContainedType != nil && inputType.ContainedType.Kind == FSKindRecord {
			// Populate column dropdown
			var options []UIDropdownOption
			for _, field := range inputType.ContainedType.Fields {
				options = append(options, UIDropdownOption{Name: field.Name, Value: field.Name})
			}
			c.colDropdown.Options = options

			if c.Column == "" && len(options) > 0 {
				c.Column = options[0].Value.(string)
			}
			if !c.colDropdown.SelectByValue(c.Column) {
				if len(options) > 0 {
					c.Column = options[0].Value.(string)
					c.colDropdown.Selected = 0
				} else {
					c.Column = ""
				}
			} else {
				c.Column = c.colDropdown.GetSelectedOption().Value.(string)
			}

			// Update output type to be Table with modified column
			newFields := make([]FlowField, len(inputType.ContainedType.Fields))
			copy(newFields, inputType.ContainedType.Fields)
			for i, f := range newFields {
				if f.Name == c.Column {
					newFields[i].Type = &FlowType{Kind: c.TargetKind}
					break
				}
			}
			n.OutputPorts[0].Type = FlowType{
				Kind: FSKindTable,
				ContainedType: &FlowType{
					Kind:   FSKindRecord,
					Fields: newFields,
				},
			}
		} else {
			// Scalar conversion
			n.OutputPorts[0].Type = FlowType{Kind: c.TargetKind}
		}
	} else {
		// Not wired, assume scalar
		n.OutputPorts[0].Type = FlowType{Kind: c.TargetKind}
	}
}

func (c *ConvertAction) UI(n *Node) {
	c.targetDropdown.Do(clay.IDI("ConvertTargetKind", n.ID), UIDropdownConfig{
		El: clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		},
		OnChange: func(before, after any) {
			c.TargetKind = after.(FlowTypeKind)
			n.Action.UpdateAndValidate(n)
		},
	})

	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{ChildGap: S1, ChildAlignment: YCENTER, Padding: clay.Padding{Top: S2}},
	}, func() {
		UIButton(clay.AUTO_ID, UIButtonConfig{
			OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
				c.IgnoreErrors = !c.IgnoreErrors
			},
		}, func() {
			UIImage(clay.AUTO_ID, util.Tern(c.IgnoreErrors, ImgToggleDown, ImgToggleRight), clay.EL{})
		})
		clay.TEXT("Ignore Errors", clay.TextElementConfig{TextColor: White})
	})

	if n.InputIsWired(0) && n.InputPorts[0].Type.Kind == FSKindTable {
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{Padding: clay.Padding{Bottom: S2}},
		}, func() {
			clay.TEXT("Column:", clay.TextElementConfig{TextColor: White})
		})
		c.colDropdown.Do(clay.IDI("ConvertColumn", n.ID), UIDropdownConfig{
			El: clay.EL{
				Layout: clay.LAY{Sizing: GROWH},
			},
			OnChange: func(before, after any) {
				c.Column = after.(string)
				n.Action.UpdateAndValidate(n)
			},
		})
	}
}

func (c *ConvertAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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

		if input.Type.Kind == FSKindTable {
			// Table conversion
			if c.Column == "" {
				res.Err = fmt.Errorf("no column selected for conversion")
				return
			}

			// Find column index
			colIndex := -1
			fields := input.Type.ContainedType.Fields
			for i, f := range fields {
				if f.Name == c.Column {
					colIndex = i
					break
				}
			}
			if colIndex == -1 {
				res.Err = fmt.Errorf("column %s not found in input table", c.Column)
				return
			}

			// Create new output type
			newFields := make([]FlowField, len(fields))
			copy(newFields, fields)
			newFields[colIndex].Type = &FlowType{Kind: c.TargetKind}
			outputType := &FlowType{
				Kind: FSKindTable,
				ContainedType: &FlowType{
					Kind:   FSKindRecord,
					Fields: newFields,
				},
			}

			// Iterate rows and convert specific column
			var newRows [][]FlowValueField
			rows := input.TableValue
			for _, row := range rows {
				// Check context periodically
				select {
				case <-ctx.Done():
					res.Err = ctx.Err()
					return
				default:
				}

				newRow := make([]FlowValueField, len(row))
				copy(newRow, row)

				// Convert the specific column
				if colIndex < len(row) {
					convertedVal, err := ConvertValue(row[colIndex].Value, c.TargetKind)
					if err != nil {
						if c.IgnoreErrors {
							convertedVal = ZeroValue(c.TargetKind)
						} else {
							res.Err = fmt.Errorf("conversion failed for value %v: %w", row[colIndex].Value, err)
							return
						}
					}
					newRow[colIndex].Value = convertedVal
				}

				newRows = append(newRows, newRow)
			}

			res = NodeActionResult{
				Outputs: []FlowValue{{
					Type:       outputType,
					TableValue: newRows,
				}},
			}

		} else {
			// Scalar conversion
			resVal, err := ConvertValue(input, c.TargetKind)
			if err != nil {
				if c.IgnoreErrors {
					resVal = ZeroValue(c.TargetKind)
				} else {
					res.Err = err
					return
				}
			}

			res = NodeActionResult{
				Outputs: []FlowValue{resVal},
			}
		}
	}()
	return done
}

func (c *ConvertAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func ZeroValue(kind FlowTypeKind) FlowValue {
	switch kind {
	case FSKindBytes:
		return NewStringValue("")
	case FSKindInt64:
		return NewInt64Value(0, 0)
	case FSKindFloat64:
		return NewFloat64Value(0, 0)
	default:
		return FlowValue{Type: &FlowType{Kind: kind}}
	}
}

func ConvertValue(v FlowValue, target FlowTypeKind) (FlowValue, error) {
	if v.Type.Kind == target {
		return v, nil
	}

	switch target {
	case FSKindBytes: // To String
		var str string
		switch v.Type.Kind {
		case FSKindBytes:
			return v, nil
		case FSKindInt64:
			str = strconv.FormatInt(v.Int64Value, 10)
		case FSKindFloat64:
			str = strconv.FormatFloat(v.Float64Value, 'f', -1, 64)
		default:
			return FlowValue{}, fmt.Errorf("cannot convert %s to Bytes", v.Type)
		}
		return NewStringValue(str), nil

	case FSKindInt64: // To Int
		var val int64
		var err error
		switch v.Type.Kind {
		case FSKindBytes:
			val, err = strconv.ParseInt(string(v.BytesValue), 10, 64)
			if err != nil {
				// Try float parsing first
				f, err2 := strconv.ParseFloat(string(v.BytesValue), 64)
				if err2 == nil {
					val = int64(f)
				} else {
					return FlowValue{}, err
				}
			}
		case FSKindInt64:
			return v, nil
		case FSKindFloat64:
			val = int64(v.Float64Value)
		default:
			return FlowValue{}, fmt.Errorf("cannot convert %s to Int64", v.Type)
		}
		return NewInt64Value(val, 0), nil

	case FSKindFloat64: // To Float
		var val float64
		var err error
		switch v.Type.Kind {
		case FSKindBytes:
			val, err = strconv.ParseFloat(string(v.BytesValue), 64)
			if err != nil {
				return FlowValue{}, err
			}
		case FSKindInt64:
			val = float64(v.Int64Value)
		case FSKindFloat64:
			return v, nil
		default:
			return FlowValue{}, fmt.Errorf("cannot convert %s to Float64", v.Type)
		}
		return NewFloat64Value(val, 0), nil
	}

	return FlowValue{}, fmt.Errorf("unsupported conversion target: %v", target)
}
