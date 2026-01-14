package nodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type AggregateAction struct {
	ops core.UIDropdown
}

var aggOptions = []core.UIDropdownOption{
	{Name: "Min", Value: AggOpMin},
	{Name: "Max", Value: AggOpMax},
	{Name: "Mean", Value: AggOpMean},
}

func NewAggregateNode(op string) *core.Node {
	action := AggregateAction{
		ops: core.UIDropdown{
			Options: aggOptions,
		},
	}
	action.ops.SelectByName(op)

	return &core.Node{
		Name: "Aggregate",

		InputPorts: []core.NodePort{{
			Name: "Input",
			Type: core.FlowType{Kind: core.FSKindAny},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Result",
			Type: core.FlowType{Kind: core.FSKindAny},
		}},

		Action: &action,
	}
}

var _ core.NodeAction = &AggregateAction{}

func (a *AggregateAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	wire, hasWire := n.GetInputWire(0)
	if hasWire {
		if core.Typecheck(wire.Type(), core.NewListType(core.FlowType{Kind: core.FSKindInt64})) == nil {
			// List[Int64] -> Int64
			n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindInt64}
		} else if core.Typecheck(wire.Type(), core.NewListType(core.FlowType{Kind: core.FSKindFloat64})) == nil {
			// List[Float64] -> Float64
			n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindFloat64}
		} else if core.Typecheck(wire.Type(), core.NewAnyTableType()) == nil {
			// Table[Any] -> Table[Any] (only numeric columns aggregated, other columns cleared)
			n.OutputPorts[0].Type = wire.Type()
		} else {
			// Dunno, catch it at runtime
			n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindAny}
		}
	} else {
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindAny}
	}

	if !hasWire {
		n.Valid = false
	}
}

func (a *AggregateAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("AggregateUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("AggregateRow1", n.ID), clay.EL{
			Layout: clay.LAY{
				Sizing:         core.GROWH,
				ChildAlignment: core.YCENTER,
			},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("AggregateSpacer", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		a.ops.Do(clay.IDI("AggregateDropdown", n.ID), core.UIDropdownConfig{
			El: clay.EL{
				Layout: clay.LAY{Sizing: core.GROWH},
			},
		})
	})
}

func (a *AggregateAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)

	go func() {
		var res core.NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		util.Assert(n.Action != nil && n.Action.Tag() == "AggregateAction", fmt.Sprintf("expected aggregate node, got %v", n))

		input, ok, err := n.GetInputValue(0)
		if !ok {
			res.Err = errors.New("an input node is required")
			return
		}
		if err != nil {
			res.Err = err
			return
		}

		op := a.ops.GetSelectedOption().Value.(AggOp)
		switch input.Type.Kind {
		case core.FSKindList:
			agged, err := op(input.ListValue, *input.Type.ContainedType)
			if err != nil {
				res.Err = err
				return
			}
			res = core.NodeActionResult{
				Outputs: []core.FlowValue{agged},
			}
		case core.FSKindTable:
			aggedRow := make([]core.FlowValueField, len(input.Type.ContainedType.Fields))
			for col, field := range input.Type.ContainedType.Fields {
				select {
				case <-ctx.Done():
					res.Err = ctx.Err()
					return
				default:
				}

				agged, err := op(input.ColumnValues(col), *field.Type)
				if err != nil {
					res.Err = fmt.Errorf("for column %s: %v", field.Name, err)
					return
				}
				aggedRow[col] = core.FlowValueField{
					Name:  field.Name,
					Value: agged,
				}
			}
			res = core.NodeActionResult{
				Outputs: []core.FlowValue{{
					Type:       input.Type,
					TableValue: [][]core.FlowValueField{aggedRow},
				}},
			}
		default:
			res.Err = fmt.Errorf("can only aggregate lists or tables, not %s", input.Type)
		}
	}()
	return done
}

func (a *AggregateAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (n *AggregateAction) Serialize(s *core.Serializer) bool {
	if s.Encode {
		s.WriteStr(n.ops.GetSelectedOption().Name)
	} else {
		selected, ok := s.ReadStr()
		if !ok {
			return false
		}
		n.ops = core.UIDropdown{Options: aggOptions}
		n.ops.SelectByName(selected)
		util.Assert(n.ops.GetSelectedOption().Name == selected, fmt.Sprintf("aggregate %s should have been selected, but %s was instead", selected, n.ops.GetSelectedOption().Name))
	}
	return s.Ok()
}

type AggOp = func(vals []core.FlowValue, t core.FlowType) (core.FlowValue, error)

var _ AggOp = AggOpMin
var _ AggOp = AggOpMax
var _ AggOp = AggOpMean

func AggOpMin(vals []core.FlowValue, t core.FlowType) (core.FlowValue, error) {
	if len(vals) == 0 {
		// Zero value of the desired type, if no values at all
		return core.FlowValue{Type: &t}, nil
	}

	switch t.Kind {
	case core.FSKindInt64:
		res := vals[0].Int64Value
		for _, v := range vals {
			res = util.Min(res, v.Int64Value)
		}
		return core.FlowValue{Type: &t, Int64Value: res}, nil
	case core.FSKindFloat64:
		res := vals[0].Float64Value
		for _, v := range vals {
			res = util.Min(res, v.Float64Value)
		}
		return core.FlowValue{Type: &t, Float64Value: res}, nil
	default:
		return core.FlowValue{}, fmt.Errorf("cannot min values of type %s", t)
	}
}

func AggOpMax(vals []core.FlowValue, t core.FlowType) (core.FlowValue, error) {
	if len(vals) == 0 {
		// Zero value of the desired type, if no values at all
		return core.FlowValue{Type: &t}, nil
	}

	switch t.Kind {
	case core.FSKindInt64:
		res := vals[0].Int64Value
		for _, v := range vals {
			res = util.Max(res, v.Int64Value)
		}
		return core.FlowValue{Type: &t, Int64Value: res}, nil
	case core.FSKindFloat64:
		res := vals[0].Float64Value
		for _, v := range vals {
			res = util.Max(res, v.Float64Value)
		}
		return core.FlowValue{Type: &t, Float64Value: res}, nil
	default:
		return core.FlowValue{}, fmt.Errorf("cannot max values of type %s", t)
	}
}

func AggOpMean(vals []core.FlowValue, t core.FlowType) (core.FlowValue, error) {
	if len(vals) == 0 {
		// Zero value of the desired type, if no values at all
		return core.FlowValue{Type: &t}, nil
	}

	switch t.Kind {
	case core.FSKindInt64:
		var sum int64
		for _, v := range vals {
			sum += v.Int64Value
		}
		return core.FlowValue{Type: &t, Int64Value: sum / int64(len(vals))}, nil
	case core.FSKindFloat64:
		var sum float64
		for _, v := range vals {
			sum += v.Float64Value
		}
		return core.FlowValue{Type: &t, Float64Value: sum / float64(len(vals))}, nil
	default:
		return core.FlowValue{}, fmt.Errorf("cannot average values of type %s", t)
	}
}
