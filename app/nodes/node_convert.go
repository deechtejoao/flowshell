package nodes

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type ConvertAction struct {
	TargetKind     core.FlowTypeKind
	TargetDropdown core.UIDropdown

	Column      string
	ColDropdown core.UIDropdown

	IgnoreErrors bool
}

var convertOptions = []core.UIDropdownOption{
	{Name: "Bytes (String)", Value: core.FSKindBytes},
	{Name: "Int64", Value: core.FSKindInt64},
	{Name: "Float64", Value: core.FSKindFloat64},
}

func NewConvertNode() *core.Node {
	action := ConvertAction{
		TargetKind: core.FSKindBytes, // Default
		TargetDropdown: core.UIDropdown{
			Options: convertOptions,
		},
		ColDropdown: core.UIDropdown{
			Options: []core.UIDropdownOption{},
		},
	}
	action.TargetDropdown.SelectByValue(core.FSKindBytes)

	return &core.Node{
		Name: "Convert Type",

		InputPorts: []core.NodePort{{
			Name: "Input",
			Type: core.FlowType{Kind: core.FSKindAny},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Output",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},

		Action: &action,
	}
}

var _ core.NodeAction = &ConvertAction{}

func (c *ConvertAction) Serialize(s *core.Serializer) bool {
	core.SInt(s, (*int)(&c.TargetKind))
	core.SStr(s, &c.Column)
	core.SBool(s, &c.IgnoreErrors)
	return s.Ok()
}

func (c *ConvertAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	// Ensure target dropdown matches state
	if c.TargetDropdown.Selected == -1 {
		c.TargetDropdown.SelectByValue(c.TargetKind)
	} else {
		c.TargetKind = c.TargetDropdown.GetSelectedOption().Value.(core.FlowTypeKind)
	}

	// Update input/output types and column dropdown
	if wire, ok := n.GetInputWire(0); ok {
		inputType := wire.Type()
		if inputType.Kind == core.FSKindTable && inputType.ContainedType != nil && inputType.ContainedType.Kind == core.FSKindRecord {
			// Populate column dropdown
			var options []core.UIDropdownOption
			for _, field := range inputType.ContainedType.Fields {
				options = append(options, core.UIDropdownOption{Name: field.Name, Value: field.Name})
			}
			c.ColDropdown.Options = options

			if c.Column == "" && len(options) > 0 {
				c.Column = options[0].Value.(string)
			}
			if !c.ColDropdown.SelectByValue(c.Column) {
				if len(options) > 0 {
					c.Column = options[0].Value.(string)
					c.ColDropdown.Selected = 0
				} else {
					c.Column = ""
				}
			} else {
				c.Column = c.ColDropdown.GetSelectedOption().Value.(string)
			}

			// Update output type to be Table with modified column
			newFields := make([]core.FlowField, len(inputType.ContainedType.Fields))
			copy(newFields, inputType.ContainedType.Fields)
			for i, f := range newFields {
				if f.Name == c.Column {
					newFields[i].Type = &core.FlowType{Kind: c.TargetKind}
					break
				}
			}
			n.OutputPorts[0].Type = core.FlowType{
				Kind: core.FSKindTable,
				ContainedType: &core.FlowType{
					Kind:   core.FSKindRecord,
					Fields: newFields,
				},
			}
		} else {
			// Scalar conversion
			n.OutputPorts[0].Type = core.FlowType{Kind: c.TargetKind}
		}
	} else {
		// Not wired, assume scalar
		n.OutputPorts[0].Type = core.FlowType{Kind: c.TargetKind}
	}
}

func (c *ConvertAction) UI(n *core.Node) {
	c.TargetDropdown.Do(clay.IDI("ConvertTargetKind", n.ID), core.UIDropdownConfig{
		El: clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH},
		},
		OnChange: func(before, after any) {
			c.TargetKind = after.(core.FlowTypeKind)
			n.Action.UpdateAndValidate(n)
		},
	})

	clay.CLAY(clay.IDI("ConvertIgnoreErrorsRow", n.ID), clay.EL{
		Layout: clay.LAY{ChildGap: core.S1, ChildAlignment: core.YCENTER, Padding: clay.Padding{Top: core.S2}},
	}, func() {
		core.UIButton(clay.IDI("ConvertIgnoreErrorsBtn", n.ID), core.UIButtonConfig{
			OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
				c.IgnoreErrors = !c.IgnoreErrors
			},
		}, func() {
			core.UIImage(clay.IDI("ConvertIgnoreErrorsIcon", n.ID), util.Tern(c.IgnoreErrors, core.ImgToggleDown, core.ImgToggleRight), clay.EL{})
		})
		clay.TEXT("Ignore Errors", clay.TextElementConfig{TextColor: core.White})
	})

	if wire, ok := n.GetInputWire(0); ok && wire.Type().Kind == core.FSKindTable {
		clay.CLAY(clay.IDI("ConvertColumnLabelRow", n.ID), clay.EL{
			Layout: clay.LAY{Padding: clay.Padding{Bottom: core.S2}},
		}, func() {
			clay.TEXT("Column:", clay.TextElementConfig{TextColor: core.White})
		})
		c.ColDropdown.Do(clay.IDI("ConvertColumn", n.ID), core.UIDropdownConfig{
			El: clay.EL{
				Layout: clay.LAY{Sizing: core.GROWH},
			},
			OnChange: func(before, after any) {
				c.Column = after.(string)
				n.Action.UpdateAndValidate(n)
			},
		})
	}

	core.UISpacer(clay.IDI("ConvertSpacer", n.ID), core.GROWH)

	// Render ports
	core.UIInputPort(n, 0)
	core.UIOutputPort(n, 0)
}

func (c *ConvertAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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

		if input.Type.Kind == core.FSKindTable {
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
			newFields := make([]core.FlowField, len(fields))
			copy(newFields, fields)
			newFields[colIndex].Type = &core.FlowType{Kind: c.TargetKind}
			outputType := &core.FlowType{
				Kind: core.FSKindTable,
				ContainedType: &core.FlowType{
					Kind:   core.FSKindRecord,
					Fields: newFields,
				},
			}

			// Iterate rows and convert specific column
			var newRows [][]core.FlowValueField
			rows := input.TableValue
			for _, row := range rows {
				// Check context periodically
				select {
				case <-ctx.Done():
					res.Err = ctx.Err()
					return
				default:
				}

				newRow := make([]core.FlowValueField, len(row))
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

			res = core.NodeActionResult{
				Outputs: []core.FlowValue{{
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

			res = core.NodeActionResult{
				Outputs: []core.FlowValue{resVal},
			}
		}
	}()
	return done
}

func (c *ConvertAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func ZeroValue(kind core.FlowTypeKind) core.FlowValue {
	switch kind {
	case core.FSKindBytes:
		return core.NewStringValue("")
	case core.FSKindInt64:
		return core.NewInt64Value(0, 0)
	case core.FSKindFloat64:
		return core.NewFloat64Value(0, 0)
	default:
		return core.FlowValue{Type: &core.FlowType{Kind: kind}}
	}
}

func ConvertValue(v core.FlowValue, target core.FlowTypeKind) (core.FlowValue, error) {
	if v.Type.Kind == target {
		return v, nil
	}

	switch target {
	case core.FSKindBytes: // To String
		var str string
		switch v.Type.Kind {
		case core.FSKindBytes:
			return v, nil
		case core.FSKindStream:
			if v.StreamValue == nil {
				return core.NewStringValue(""), nil
			}
			b, err := io.ReadAll(v.StreamValue)
			if err != nil {
				return core.FlowValue{}, err
			}
			str = string(b)
		case core.FSKindInt64:
			str = strconv.FormatInt(v.Int64Value, 10)
		case core.FSKindFloat64:
			str = strconv.FormatFloat(v.Float64Value, 'f', -1, 64)
		default:
			return core.FlowValue{}, fmt.Errorf("cannot convert %s to Bytes", v.Type)
		}
		return core.NewStringValue(str), nil

	case core.FSKindInt64: // To Int
		var val int64
		var err error
		switch v.Type.Kind {
		case core.FSKindBytes:
			val, err = strconv.ParseInt(string(v.BytesValue), 10, 64)
			if err != nil {
				// Try float parsing first
				f, err2 := strconv.ParseFloat(string(v.BytesValue), 64)
				if err2 == nil {
					val = int64(f)
				} else {
					return core.FlowValue{}, err
				}
			}
		case core.FSKindStream:
			if v.StreamValue == nil {
				val = 0
			} else {
				b, err2 := io.ReadAll(v.StreamValue)
				if err2 != nil {
					return core.FlowValue{}, err2
				}
				val, err = strconv.ParseInt(string(b), 10, 64)
				if err != nil {
					f, err3 := strconv.ParseFloat(string(b), 64)
					if err3 == nil {
						val = int64(f)
					} else {
						return core.FlowValue{}, err
					}
				}
			}
		case core.FSKindInt64:
			return v, nil
		case core.FSKindFloat64:
			val = int64(v.Float64Value)
		default:
			return core.FlowValue{}, fmt.Errorf("cannot convert %s to Int64", v.Type)
		}
		return core.NewInt64Value(val, 0), nil

	case core.FSKindFloat64: // To Float
		var val float64
		var err error
		switch v.Type.Kind {
		case core.FSKindBytes:
			val, err = strconv.ParseFloat(string(v.BytesValue), 64)
			if err != nil {
				return core.FlowValue{}, err
			}
		case core.FSKindStream:
			if v.StreamValue == nil {
				val = 0
			} else {
				b, err2 := io.ReadAll(v.StreamValue)
				if err2 != nil {
					return core.FlowValue{}, err2
				}
				val, err = strconv.ParseFloat(string(b), 64)
				if err != nil {
					return core.FlowValue{}, err
				}
			}
		case core.FSKindInt64:
			val = float64(v.Int64Value)
		case core.FSKindFloat64:
			return v, nil
		default:
			return core.FlowValue{}, fmt.Errorf("cannot convert %s to Float64", v.Type)
		}
		return core.NewFloat64Value(val, 0), nil
	}

	return core.FlowValue{}, fmt.Errorf("unsupported conversion target: %v", target)
}