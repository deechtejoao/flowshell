package app

import (
	"context"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type PromptUserAction struct {
	Title        string
	Message      string
	DefaultValue string
}

func NewPromptUserNode() *Node {
	return &Node{
		Name: "Prompt User",
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindBytes}}, // Output the entered text
		},
		Action: &PromptUserAction{
			Title:   "Input Required",
			Message: "Please enter a value:",
		},
	}
}

var _ NodeAction = &PromptUserAction{}

func (a *PromptUserAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *PromptUserAction) UI(n *Node) {
	clay.CLAY(clay.IDI("PromptUser", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.TEXT("Title:", clay.TextElementConfig{TextColor: White})
		UITextBox(clay.IDI("PromptTitle", n.ID), &a.Title, UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
		})

		clay.TEXT("Message:", clay.TextElementConfig{TextColor: White})
		UITextBox(clay.IDI("PromptMessage", n.ID), &a.Message, UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
		})

		clay.TEXT("Default:", clay.TextElementConfig{TextColor: White})
		UITextBox(clay.IDI("PromptDefault", n.ID), &a.DefaultValue, UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
		})

		UIOutputPort(n, 0)
	})
}

func (a *PromptUserAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *PromptUserAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)

	go func() {
		defer close(done)

		// Call the global prompt handler
		val, err := RequestPrompt(ctx, a.Title, a.Message, a.DefaultValue)
		if err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{NewStringValue(val)},
		}
	}()

	return done
}

func (a *PromptUserAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Title)
	SStr(s, &a.Message)
	SStr(s, &a.DefaultValue)
	return s.Ok()
}
