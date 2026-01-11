package app

import (
	"fmt"
	"strconv"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ConvertAction struct {
	TargetKind FlowTypeKind
	dropdown   UIDropdown
}

var convertOptions = []UIDropdownOption{
	{Name: "Bytes (String)", Value: FSKindBytes},
	{Name: "Int64", Value: FSKindInt64},
	{Name: "Float64", Value: FSKindFloat64},
}

func NewConvertNode() *Node {
	action := ConvertAction{
		TargetKind: FSKindBytes, // Default
		dropdown: UIDropdown{
			Options: convertOptions,
		},
	}
	action.dropdown.SelectByValue(FSKindBytes)

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

func (c *ConvertAction) Tag() string {
	return "ConvertAction"
}

func (c *ConvertAction) Serialize(s *Serializer) bool {
	SInt(s, (*int)(&c.TargetKind))
	return s.Ok()
}

func (c *ConvertAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	// Ensure dropdown matches state (if loaded from save)
	if c.dropdown.Selected == -1 {
		c.dropdown.SelectByValue(c.TargetKind)
	} else {
		c.TargetKind = c.dropdown.GetSelectedOption().Value.(FlowTypeKind)
	}

	n.OutputPorts[0].Type = FlowType{Kind: c.TargetKind}
}

func (c *ConvertAction) UI(n *Node) {
	c.dropdown.Do(clay.IDI("ConvertTargetKind", n.ID), UIDropdownConfig{
		El: clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		},
		OnChange: func(before, after any) {
			c.TargetKind = after.(FlowTypeKind)
			n.Action.UpdateAndValidate(n)
		},
	})
}

func (c *ConvertAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		defer close(done)

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		resVal, err := ConvertValue(input, c.TargetKind)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{resVal},
		}
	}()
	return done
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
