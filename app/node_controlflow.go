package app

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// --- Gate Node ---

// GEN:NodeAction
type GateAction struct{}

func NewGateNode() *Node {
	return &Node{
		Name: "Gate",
		InputPorts: []NodePort{
			{Name: "Value", Type: FlowType{Kind: FSKindAny}},
			{Name: "Condition", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Filtered", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &GateAction{},
	}
}

var _ NodeAction = &GateAction{}

func (a *GateAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *GateAction) UI(n *Node) {
	clay.CLAY(clay.IDI("GateNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.TEXT("Passes 'Value' if 'Condition' is true", clay.TextElementConfig{TextColor: LightGray})

		UIInputPort(n, 0)
		UIInputPort(n, 1)
		UIOutputPort(n, 0)
	})
}

func (a *GateAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GateAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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
			done <- NodeActionResult{Err: fmt.Errorf("missing inputs")}
			return
		}

		// Logic:
		// If Condition is truthy -> Output Value
		// If Condition is false -> Output Skipped Value

		isTrue := IsTruthy(condition)
		output := value
		if !isTrue {
			output.Skipped = true
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{output},
		}
	}()

	return done
}

func (a *GateAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// --- Merge Node ---

// GEN:NodeAction
type MergeAction struct{}

func NewMergeNode() *Node {
	return &Node{
		Name:                "Merge",
		HandleSkippedInputs: true,
		InputPorts: []NodePort{
			{Name: "A", Type: FlowType{Kind: FSKindAny}},
			{Name: "B", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Merged", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &MergeAction{},
	}
}

var _ NodeAction = &MergeAction{}

func (a *MergeAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *MergeAction) UI(n *Node) {
	clay.CLAY(clay.IDI("MergeNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.TEXT("Passes first non-skipped input", clay.TextElementConfig{TextColor: LightGray})

		UIInputPort(n, 0)
		UIInputPort(n, 1)
		UIOutputPort(n, 0)
	})
}

func (a *MergeAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MergeAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		valA, okA, errA := n.GetInputValue(0)
		valB, okB, errB := n.GetInputValue(1)

		if errA != nil {
			done <- NodeActionResult{Err: errA}
			return
		}
		if errB != nil {
			done <- NodeActionResult{Err: errB}
			return
		}

		// Strategy:
		// If A is valid and not skipped -> A
		// Else if B is valid and not skipped -> B
		// Else -> Skipped (using A's type or generic)

		var output FlowValue

		if okA && !valA.Skipped {
			output = valA
		} else if okB && !valB.Skipped {
			output = valB
		} else {
			// Both skipped or missing.
			// Return skipped value. Prefer A's type if available, else Any.
			if okA {
				output = valA
			} else if okB {
				output = valB
			} else {
				// No inputs connected?
				// Just return Generic skipped
				output = FlowValue{Type: &FlowType{Kind: FSKindAny}, Skipped: true}
			}
			output.Skipped = true
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{output},
		}
	}()

	return done
}

func (a *MergeAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
