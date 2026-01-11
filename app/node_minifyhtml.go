package app

import (
	"regexp"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type MinifyHTMLAction struct{}

func NewMinifyHTMLNode() *Node {
	return &Node{
		ID:   NewNodeID(),
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

func (a *MinifyHTMLAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		defer close(done)

		input, ok, err := n.GetInputValue(0)
		if !ok || err != nil {
			done <- NodeActionResult{Err: err}
			return
		}

		html := string(input.BytesValue)
		// Simple minification: replace >\s+< with ><
		re := regexp.MustCompile(`>\s+<`)
		minified := re.ReplaceAllString(html, "><")

		done <- NodeActionResult{
			Outputs: []FlowValue{{
				Type:       &FlowType{Kind: FSKindBytes},
				BytesValue: []byte(minified),
			}},
		}
	}()
	return done
}
