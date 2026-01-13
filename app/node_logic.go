package app

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type IfElseAction struct{}

func NewIfElseNode() *Node {
	return &Node{
		Name: "If / Else",
		InputPorts: []NodePort{
			{Name: "Value", Type: FlowType{Kind: FSKindAny}},
			{Name: "Condition", Type: FlowType{Kind: FSKindAny}}, // Accepts Any, checks for Truthy
		},
		OutputPorts: []NodePort{
			{Name: "True", Type: FlowType{Kind: FSKindAny}},
			{Name: "False", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &IfElseAction{},
	}
}

var _ NodeAction = &IfElseAction{}

func (a *IfElseAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *IfElseAction) UI(n *Node) {
	clay.CLAY(clay.IDI("IfElseNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.TEXT("Routes 'Value' based on 'Condition'", clay.TextElementConfig{TextColor: LightGray})

		UIInputPort(n, 0)
		UIInputPort(n, 1)
		UIOutputPort(n, 0)
		UIOutputPort(n, 1)
	})
}

func (a *IfElseAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *IfElseAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		value, ok1, err1 := n.GetInputValue(0)
		condition, ok2, err2 := n.GetInputValue(1)

		if err1 != nil {
			done <- NodeActionResult{Err: err1}
			return
		}
		if err2 != nil {
			done <- NodeActionResult{Err: err2}
			return
		}

		if !ok1 || !ok2 {
			// Should verify inputs are available, but engine checks existence.
			// If missing connection, we error? Or skip?
			// Usually engine handles dependency wait.
			// If GetInputValue returns false, it means no wire connected or no value produced?
			// Actually, engine waits for inputs. GetInputValue checks wire.
			// If no wire connected, we can't run.
			done <- NodeActionResult{Err: fmt.Errorf("Missing inputs")}
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
		// Node says OutputPorts are Any. So we return the value as is.
		// Wait, output ports adapt to input? No, generic "Any" passes through.

		done <- NodeActionResult{
			Outputs: []FlowValue{trueOut, falseOut},
		}
	}()

	return done
}

func (a *IfElseAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

func IsTruthy(v FlowValue) bool {
	if v.Skipped {
		return false
	}

	switch v.Type.Kind {
	case FSKindInt64:
		return v.Int64Value != 0
	case FSKindFloat64:
		return v.Float64Value != 0
	case FSKindBytes:
		return len(v.BytesValue) > 0 // Empty string is false?
	case FSKindList:
		return len(v.ListValue) > 0
	case FSKindTable:
		return len(v.TableValue) > 0
	}
	return true // Default true for other types?
}
