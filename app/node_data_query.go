package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/bvisness/flowshell/clay"
	"github.com/tidwall/gjson"
)

// GEN:NodeAction
type JsonQueryAction struct {
	Query string
}

func NewJsonQueryNode() *Node {
	return &Node{
		Name: "JSON Query",
		Action: &JsonQueryAction{
			Query: "foo.bar",
		},
		InputPorts: []NodePort{
			{Name: "JSON", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindAny}},
		},
	}
}

func (a *JsonQueryAction) UpdateAndValidate(n *Node) { n.Valid = true }

func (a *JsonQueryAction) UI(n *Node) {
	clay.CLAY(clay.IDI("JsonQUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("Spacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			clay.TEXT("GJSON Query:", clay.TextElementConfig{FontSize: 12, TextColor: LightGray})
		})

		clay.CLAY(clay.IDI("Row3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("QueryStr", n.ID), &a.Query, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})
	})
}

func (a *JsonQueryAction) Run(n *Node) <-chan NodeActionResult {
	ch := make(chan NodeActionResult, 1)
	go func() {
		defer close(ch)
		inputVal, ok, err := n.GetInputValue(0)
		if err != nil {
			ch <- NodeActionResult{Err: err}
			return
		}
		if !ok {
			ch <- NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		jsonStr := string(inputVal.BytesValue)
		res := gjson.Get(jsonStr, a.Query)
		var output FlowValue

		switch res.Type {
		case gjson.String:
			output = NewStringValue(res.String())
		case gjson.Number:
			output = NewFloat64Value(res.Float(), 0)
		case gjson.True:
			output = NewInt64Value(1, 0)
		case gjson.False:
			output = NewInt64Value(0, 0)
		case gjson.Null:
			output = FlowValue{Type: &FlowType{Kind: FSKindAny}}
		default: // JSON (Object/Array) -> treat as String
			output = NewStringValue(res.String())
		}

		ch <- NodeActionResult{Outputs: []FlowValue{output}}
	}()
	return ch
}

func (a *JsonQueryAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *JsonQueryAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Query)
	return s.Ok()
}

// GEN:NodeAction
type XmlQueryAction struct {
	XPath string
}

func NewXmlQueryNode() *Node {
	return &Node{
		Name: "XML Query",
		Action: &XmlQueryAction{
			XPath: "//item",
		},
		InputPorts: []NodePort{
			{Name: "XML", Type: FlowType{Kind: FSKindBytes}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindBytes}},
		},
	}
}

func (a *XmlQueryAction) UpdateAndValidate(n *Node) { n.Valid = true }

func (a *XmlQueryAction) UI(n *Node) {
	clay.CLAY(clay.IDI("XmlQUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("Spacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			clay.TEXT("XPath:", clay.TextElementConfig{FontSize: 12, TextColor: LightGray})
		})

		clay.CLAY(clay.IDI("Row3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("XPathStr", n.ID), &a.XPath, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})
	})
}

func (a *XmlQueryAction) Run(n *Node) <-chan NodeActionResult {
	ch := make(chan NodeActionResult, 1)
	go func() {
		defer close(ch)
		inputVal, ok, err := n.GetInputValue(0)
		if err != nil {
			ch <- NodeActionResult{Err: err}
			return
		}
		if !ok {
			ch <- NodeActionResult{Err: fmt.Errorf("missing input")}
			return
		}

		xmlStr := string(inputVal.BytesValue)
		doc, err := xmlquery.Parse(strings.NewReader(xmlStr))
		if err != nil {
			ch <- NodeActionResult{Err: fmt.Errorf("failed to parse XML: %v", err)}
			return
		}

		list, err := xmlquery.QueryAll(doc, a.XPath)
		if err != nil {
			ch <- NodeActionResult{Err: fmt.Errorf("xpath query failed: %v", err)}
			return
		}

		var builder strings.Builder
		for i, n := range list {
			if i > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(n.OutputXML(true))
		}

		ch <- NodeActionResult{Outputs: []FlowValue{NewStringValue(builder.String())}}
	}()
	return ch
}

func (a *XmlQueryAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *XmlQueryAction) Serialize(s *Serializer) bool {
	SStr(s, &a.XPath)
	return s.Ok()
}
