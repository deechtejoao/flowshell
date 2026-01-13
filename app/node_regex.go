package app

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type RegexMatchAction struct{}

func NewRegexMatchNode() *Node {
	return &Node{
		Name: "Regex Match",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Pattern", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Match", Type: FlowType{Kind: FSKindInt64}}, // 0 or 1 for bool
		},
		Action: &RegexMatchAction{},
	}
}

var _ NodeAction = &RegexMatchAction{}

func (a *RegexMatchAction) UpdateAndValidate(n *Node) {
	n.Valid = true // Validation could check if pattern compiles
}

func (a *RegexMatchAction) UI(n *Node) {
	clay.CLAY(clay.IDI("RegexMatch", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 1)
		})
	})
}

func (a *RegexMatchAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val0, _, err0 := n.GetInputValue(0)
		val1, _, err1 := n.GetInputValue(1)

		if err0 != nil || err1 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v", err0, err1)}
			return
		}

		text := string(val0.BytesValue)
		pattern := string(val1.BytesValue)

		re, err := regexp.Compile(pattern)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		match := re.MatchString(text)
		val := int64(0)
		if match {
			val = 1
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewInt64Value(val, 0),
			},
		}
	}()
	return done
}

func (a *RegexMatchAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *RegexMatchAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type RegexFindAllAction struct{}

func NewRegexFindAllNode() *Node {
	return &Node{
		Name: "Regex Find All",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Pattern", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Matches", Type: NewListType(FlowType{Kind: FSKindBytes})},
		},
		Action: &RegexFindAllAction{},
	}
}

var _ NodeAction = &RegexFindAllAction{}

func (a *RegexFindAllAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *RegexFindAllAction) UI(n *Node) {
	clay.CLAY(clay.IDI("RegexFindAll", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 1)
		})
	})
}

func (a *RegexFindAllAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val0, _, err0 := n.GetInputValue(0)
		val1, _, err1 := n.GetInputValue(1)

		if err0 != nil || err1 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v", err0, err1)}
			return
		}

		text := string(val0.BytesValue)
		pattern := string(val1.BytesValue)

		re, err := regexp.Compile(pattern)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		matches := re.FindAllString(text, -1)
		var flowMatches []FlowValue
		for _, m := range matches {
			flowMatches = append(flowMatches, NewStringValue(m))
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewListValue(FlowType{Kind: FSKindBytes}, flowMatches),
			},
		}
	}()
	return done
}

func (a *RegexFindAllAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *RegexFindAllAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type RegexReplaceAction struct{}

func NewRegexReplaceNode() *Node {
	return &Node{
		Name: "Regex Replace",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Pattern", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Replacement", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &RegexReplaceAction{},
	}
}

var _ NodeAction = &RegexReplaceAction{}

func (a *RegexReplaceAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *RegexReplaceAction) UI(n *Node) {
	clay.CLAY(clay.IDI("RegexReplace", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 1)
		})
		clay.CLAY(clay.IDI("Row3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 2)
		})
	})
}

func (a *RegexReplaceAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val0, _, err0 := n.GetInputValue(0)
		val1, _, err1 := n.GetInputValue(1)
		val2, _, err2 := n.GetInputValue(2)

		if err0 != nil || err1 != nil || err2 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v, %v", err0, err1, err2)}
			return
		}

		text := string(val0.BytesValue)
		pattern := string(val1.BytesValue)
		repl := string(val2.BytesValue)

		re, err := regexp.Compile(pattern)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		result := re.ReplaceAllString(text, repl)

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewStringValue(result),
			},
		}
	}()
	return done
}

func (a *RegexReplaceAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *RegexReplaceAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type RegexSplitAction struct{}

func NewRegexSplitNode() *Node {
	return &Node{
		Name: "Regex Split",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Pattern", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Parts", Type: NewListType(FlowType{Kind: FSKindBytes})},
		},
		Action: &RegexSplitAction{},
	}
}

var _ NodeAction = &RegexSplitAction{}

func (a *RegexSplitAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *RegexSplitAction) UI(n *Node) {
	clay.CLAY(clay.IDI("RegexSplit", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 1)
		})
	})
}

func (a *RegexSplitAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val0, _, err0 := n.GetInputValue(0)
		val1, _, err1 := n.GetInputValue(1)

		if err0 != nil || err1 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v", err0, err1)}
			return
		}

		text := string(val0.BytesValue)
		pattern := string(val1.BytesValue)

		re, err := regexp.Compile(pattern)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		parts := re.Split(text, -1)
		var flowParts []FlowValue
		for _, p := range parts {
			flowParts = append(flowParts, NewStringValue(p))
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewListValue(FlowType{Kind: FSKindBytes}, flowParts),
			},
		}
	}()
	return done
}

func (a *RegexSplitAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *RegexSplitAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

