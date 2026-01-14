package nodes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type MinifyHTMLAction struct{}

func NewMinifyHTMLNode() *core.Node {
	return &core.Node{
		Name: "Minify HTML",

		InputPorts: []core.NodePort{{
			Name: "HTML",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Minified",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},

		Action: &MinifyHTMLAction{},
	}
}

var _ core.NodeAction = &MinifyHTMLAction{}

func (a *MinifyHTMLAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

func (a *MinifyHTMLAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if !wired {
		n.Valid = false
	} else if input.Type().Kind != core.FSKindBytes {
		n.Valid = false
	}
}

func (a *MinifyHTMLAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("MinifyHTMLUI", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:         core.GROWH,
			ChildAlignment: core.YCENTER,
		},
	}, func() {
		core.UIInputPort(n, 0)
		core.UISpacer(clay.IDI("MinifyHTMLSpacer", n.ID), core.GROWH)
		core.UIOutputPort(n, 0)
	})
}

func (a *MinifyHTMLAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)
	go func() {
		var res core.NodeActionResult
		defer func() {
			if r := recover(); r != nil {
				res = core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
			close(done)
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			res.Err = err
			return
		}

		html := string(input.BytesValue)
		// Simple minification: replace >\s+< with ><
		re := regexp.MustCompile(`>\s+<`)
		minified := re.ReplaceAllString(html, "><")

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:       &core.FlowType{Kind: core.FSKindBytes},
				BytesValue: []byte(minified),
			}},
		}
	}()
	return done
}

func (a *MinifyHTMLAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}
