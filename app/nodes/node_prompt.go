package nodes

import (
	"context"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type PromptUserAction struct {
	Title        string
	Message      string
	DefaultValue string
}

func NewPromptUserNode() *core.Node {
	return &core.Node{
		Name: "Prompt User",
		OutputPorts: []core.NodePort{
			{Name: "Result", Type: core.FlowType{Kind: core.FSKindBytes}}, // Output the entered text
		},
		Action: &PromptUserAction{
			Title:   "Input Required",
			Message: "Please enter a value:",
		},
	}
}

var _ core.NodeAction = &PromptUserAction{}

func (a *PromptUserAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *PromptUserAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("PromptUser", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.TEXT("Title:", clay.TextElementConfig{TextColor: core.White})
		core.UITextBox(clay.IDI("PromptTitle", n.ID), &a.Title, core.UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
		})

		clay.TEXT("Message:", clay.TextElementConfig{TextColor: core.White})
		core.UITextBox(clay.IDI("PromptMessage", n.ID), &a.Message, core.UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
		})

		clay.TEXT("Default:", clay.TextElementConfig{TextColor: core.White})
		core.UITextBox(clay.IDI("PromptDefault", n.ID), &a.DefaultValue, core.UITextBoxConfig{
			El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
		})

		core.UIOutputPort(n, 0)
	})
}

func (a *PromptUserAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *PromptUserAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)

	go func() {
		defer close(done)

		// Call the global prompt handler
		val, err := core.RequestPrompt(ctx, a.Title, a.Message, a.DefaultValue)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{core.NewStringValue(val)},
		}
	}()

	return done
}

func (a *PromptUserAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &a.Title)
	core.SStr(s, &a.Message)
	core.SStr(s, &a.DefaultValue)
	return s.Ok()
}