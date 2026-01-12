package app

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type MinifyHTMLAction struct{}

func NewMinifyHTMLNode() *Node {
	return &Node{
		Name: "Minify HTML",

		InputPorts: []NodePort{{
			Name: "HTML",
			Type: FlowType{Kind: FSKindBytes},
		}},
		OutputPorts: []NodePort{{
			Name: "Minified",
			Type: FlowType{Kind: FSKindBytes},
		}},

		Action: &MinifyHTMLAction{},
	}
}

var _ NodeAction = &MinifyHTMLAction{}

func (a *MinifyHTMLAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

func (a *MinifyHTMLAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	input, wired := n.GetInputWire(0)
	if !wired {
		n.Valid = false
	} else if input.Type().Kind != FSKindBytes {
		n.Valid = false
	}
}

func (a *MinifyHTMLAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		UIInputPort(n, 0)
		UISpacer(clay.AUTO_ID, GROWH)
		UIOutputPort(n, 0)
	})
}

func (a *MinifyHTMLAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		var res NodeActionResult
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
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

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:       &FlowType{Kind: FSKindBytes},
				BytesValue: []byte(minified),
			}},
		}
	}()
	return done
}

func (a *MinifyHTMLAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}
