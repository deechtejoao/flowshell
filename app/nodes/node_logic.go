package nodes

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type IfElseAction struct{}

func NewIfElseNode() *core.Node {
	return &core.Node{
		Name: "If / Else",
		InputPorts: []core.NodePort{
			{Name: "Value", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "Condition", Type: core.FlowType{Kind: core.FSKindAny}}, // Accepts Any, checks for Truthy
		},
		OutputPorts: []core.NodePort{
			{Name: "True", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "False", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &IfElseAction{},
	}
}

var _ core.NodeAction = &IfElseAction{}

func (a *IfElseAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *IfElseAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("NodeContent", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.TEXT("Routes 'Value' based on 'Condition'", clay.TextElementConfig{TextColor: core.LightGray})

		core.UIInputPort(n, 0)
		core.UIInputPort(n, 1)
		core.UIOutputPort(n, 0)
		core.UIOutputPort(n, 1)
	})
}

func (a *IfElseAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *IfElseAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		value, ok1, err1 := n.GetInputValue(0)
		condition, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- core.NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- core.NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			// Should verify inputs are available, but engine checks existence.
			// If missing connection, we error? Or skip?
			// Usually engine handles dependency wait.
			// If GetInputValue returns false, it means no wire connected or no value produced?
			// Actually, engine waits for inputs. GetInputValue checks wire.
			// If no wire connected, we can't run.
			done <- core.NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		// Determine truthiness
		isTrue := IsTruthy(condition)

		trueOut := value
		falseOut := value

		if isTrue {
			falseOut.Skipped = true
		} else {
			trueOut.Skipped = true
		}

		// We need to return specific types?
		// core.Node says OutputPorts are Any. So we return the value as is.
		// Wait, output ports adapt to input? No, generic "Any" passes through.

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{trueOut, falseOut},
		}
	}()

	return done
}

func (a *IfElseAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

func IsTruthy(v core.FlowValue) bool {
	if v.Skipped {
		return false
	}

	switch v.Type.Kind {
	case core.FSKindInt64:
		return v.Int64Value != 0
	case core.FSKindFloat64:
		return v.Float64Value != 0
	case core.FSKindBytes:
		return len(v.BytesValue) > 0 // Empty string is false?
	case core.FSKindList:
		return len(v.ListValue) > 0
	case core.FSKindTable:
		return len(v.TableValue) > 0
	}
	return true // Default true for other types?
}
