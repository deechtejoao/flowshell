package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// --- Join Text ---

// GEN:NodeAction
type JoinTextAction struct{}

func NewJoinTextNode() *Node {
	return &Node{
		Name: "Join Text",
		InputPorts: []NodePort{
			{Name: "List", Type: FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindBytes}}},
			{Name: "Separator", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &JoinTextAction{},
	}
}

var _ NodeAction = &JoinTextAction{}

func (a *JoinTextAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *JoinTextAction) UI(n *Node) {
	clay.CLAY(clay.IDI("JoinText", n.ID), clay.EL{
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

func (a *JoinTextAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		valList, _, err1 := n.GetInputValue(0)
		valSep, _, err2 := n.GetInputValue(1)

		if err1 != nil || err2 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v", err1, err2)}
			return
		}

		sep := ""
		if valSep.Type.Kind == FSKindBytes {
			sep = string(valSep.BytesValue)
		}

		var parts []string
		if valList.Type.Kind == FSKindList {
			for _, item := range valList.ListValue {
				if item.Type.Kind == FSKindBytes {
					parts = append(parts, string(item.BytesValue))
				} else {
					// Best effort conversion
					parts = append(parts, fmt.Sprintf("%v", FlowValueToNative(item)))
				}
			}
		}

		result := strings.Join(parts, sep)

		done <- NodeActionResult{
			Outputs: []FlowValue{NewStringValue(result)},
		}
	}()
	return done
}

func (a *JoinTextAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *JoinTextAction) Tag() string                  { return "JoinText" }
func (a *JoinTextAction) Serialize(s *Serializer) bool { return s.Ok() }

// --- Split Text ---

// GEN:NodeAction
type SplitTextAction struct{}

func NewSplitTextNode() *Node {
	return &Node{
		Name: "Split Text",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Separator", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "List", Type: NewListType(FlowType{Kind: FSKindBytes})},
		},
		Action: &SplitTextAction{},
	}
}

var _ NodeAction = &SplitTextAction{}

func (a *SplitTextAction) UpdateAndValidate(n *Node) { n.Valid = true }
func (a *SplitTextAction) UI(n *Node) {
	clay.CLAY(clay.IDI("SplitText", n.ID), clay.EL{
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

func (a *SplitTextAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		valText, _, err1 := n.GetInputValue(0)
		valSep, _, err2 := n.GetInputValue(1)

		if err1 != nil || err2 != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad inputs: %v, %v", err1, err2)}
			return
		}

		text := string(valText.BytesValue)
		sep := string(valSep.BytesValue)

		parts := strings.Split(text, sep)
		var list []FlowValue
		for _, p := range parts {
			list = append(list, NewStringValue(p))
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{NewListValue(FlowType{Kind: FSKindBytes}, list)},
		}
	}()
	return done
}

func (a *SplitTextAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *SplitTextAction) Tag() string                  { return "SplitText" }
func (a *SplitTextAction) Serialize(s *Serializer) bool { return s.Ok() }

// --- Case Convert ---

type CaseMode int

const (
	CaseUpper CaseMode = iota
	CaseLower
	CaseTitle
)

// GEN:NodeAction
type CaseConvertAction struct {
	Mode CaseMode
}

func NewCaseConvertNode() *Node {
	return &Node{
		Name: "Change Case",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &CaseConvertAction{Mode: CaseUpper},
	}
}

var _ NodeAction = &CaseConvertAction{}

func (a *CaseConvertAction) UpdateAndValidate(n *Node) { n.Valid = true }

func (a *CaseConvertAction) UI(n *Node) {
	clay.CLAY(clay.IDI("CaseUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("CaseRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("CaseSpacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		// Mode Selection
		clay.CLAY(clay.IDI("CaseRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildGap: S2},
		}, func() {
			modes := []struct {
				L string
				M CaseMode
			}{
				{"UPPER", CaseUpper}, {"lower", CaseLower}, {"Title", CaseTitle},
			}
			for _, m := range modes {
				active := a.Mode == m.M
				UIButton(clay.IDI("CaseBtn"+m.L, n.ID), UIButtonConfig{
					OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
						a.Mode = m.M
					},
					El: clay.EL{BackgroundColor: util.Tern(active, Blue, Charcoal)},
				}, func() {
					clay.TEXT(m.L, clay.TextElementConfig{TextColor: White, FontSize: 14})
				})
			}
		})
	})
}

func (a *CaseConvertAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		valText, _, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		text := string(valText.BytesValue)
		var res string
		switch a.Mode {
		case CaseUpper:
			res = strings.ToUpper(text)
		case CaseLower:
			res = strings.ToLower(text)
		case CaseTitle:
			res = cases.Title(language.English).String(text)
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{NewStringValue(res)},
		}
	}()
	return done
}

func (a *CaseConvertAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *CaseConvertAction) Tag() string { return "CaseConvert" }
func (a *CaseConvertAction) Serialize(s *Serializer) bool {
	SInt(s, &a.Mode)
	return s.Ok()
}

// --- Format String ---

// GEN:NodeAction
type FormatStringAction struct {
	Format string
}

func NewFormatStringNode() *Node {
	return &Node{
		Name: "Format String",
		InputPorts: []NodePort{
			{Name: "Value", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &FormatStringAction{Format: "Val: %v"},
	}
}

var _ NodeAction = &FormatStringAction{}

func (a *FormatStringAction) UpdateAndValidate(n *Node) { n.Valid = true }

func (a *FormatStringAction) UI(n *Node) {
	clay.CLAY(clay.IDI("FmtUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("FmtRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("FmtSpacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("FmtRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("FormatStr", n.ID), &a.Format, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})
	})
}

func (a *FormatStringAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val, _, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		native := FlowValueToNative(val)
		res := fmt.Sprintf(a.Format, native)

		done <- NodeActionResult{
			Outputs: []FlowValue{NewStringValue(res)},
		}
	}()
	return done
}

func (a *FormatStringAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *FormatStringAction) Tag() string { return "FormatString" }
func (a *FormatStringAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Format)
	return s.Ok()
}
