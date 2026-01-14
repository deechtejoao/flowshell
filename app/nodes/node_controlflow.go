package nodes

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// --- Gate core.Node ---

// GEN:NodeAction
type GateAction struct{}

func NewGateNode() *core.Node {
	return &core.Node{
		Name: "Gate",
		InputPorts: []core.NodePort{
			{Name: "Value", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "Condition", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Filtered", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &GateAction{},
	}
}

var _ core.NodeAction = &GateAction{}

func (a *GateAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *GateAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("GateNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.TEXT("Passes 'Value' if 'Condition' is true", clay.TextElementConfig{TextColor: core.LightGray})

		core.UIInputPort(n, 0)
		core.UIInputPort(n, 1)
		core.UIOutputPort(n, 0)
	})
}

func (a *GateAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GateAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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
			done <- core.NodeActionResult{Err: fmt.Errorf("missing inputs")}
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

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{output},
		}
	}()

	return done
}

func (a *GateAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

// --- Merge core.Node ---

// GEN:NodeAction
type MergeAction struct{}

func NewMergeNode() *core.Node {
	return &core.Node{
		Name:                "Merge",
		HandleSkippedInputs: true,
		InputPorts: []core.NodePort{
			{Name: "A", Type: core.FlowType{Kind: core.FSKindAny}},
			{Name: "B", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Merged", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &MergeAction{},
	}
}

var _ core.NodeAction = &MergeAction{}

func (a *MergeAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *MergeAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("MergeNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.TEXT("Passes first non-skipped input", clay.TextElementConfig{TextColor: core.LightGray})

		core.UIInputPort(n, 0)
		core.UIInputPort(n, 1)
		core.UIOutputPort(n, 0)
	})
}

func (a *MergeAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MergeAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		valA, okA, errA := n.GetInputValue(0)
		valB, okB, errB := n.GetInputValue(1)

		if errA != nil {
			done <- core.NodeActionResult{Err: errA}
			return
		}
		if errB != nil {
			done <- core.NodeActionResult{Err: errB}
			return
		}

		// Strategy:
		// If A is valid and not skipped -> A
		// Else if B is valid and not skipped -> B
		// Else -> Skipped (using A's type or generic)

		var output core.FlowValue

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
				output = core.FlowValue{Type: &core.FlowType{Kind: core.FSKindAny}, Skipped: true}
			}
			output.Skipped = true
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{output},
		}
	}()

	return done
}

func (a *MergeAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}
