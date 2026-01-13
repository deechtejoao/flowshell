package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type HTTPRequestAction struct{}

func NewHTTPRequestNode() *Node {
	return &Node{
		Name: "HTTP Request",
		InputPorts: []NodePort{
			{Name: "URL", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Method", Type: FlowType{Kind: FSKindBytes}}, // Optional, default GET
			{Name: "Body", Type: FlowType{Kind: FSKindBytes}},   // Optional
		},
		OutputPorts: []NodePort{
			{Name: "Status", Type: FlowType{Kind: FSKindInt64}},
			{Name: "Body", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &HTTPRequestAction{},
	}
}

var _ NodeAction = &HTTPRequestAction{}

func (a *HTTPRequestAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *HTTPRequestAction) UI(n *Node) {
	clay.CLAY(clay.IDI("HTTPRequest", n.ID), clay.EL{
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
			UISpacer(clay.AUTO_ID, GROWH)
			UIOutputPort(n, 1)
		})
		clay.CLAY(clay.IDI("Row3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 2)
		})
	})
}

func (a *HTTPRequestAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		// Get Inputs
		// URL is required (index 0)
		valURL, _, errURL := n.GetInputValue(0)
		if errURL != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad URL input: %v", errURL)}
			return
		}
		urlStr := string(valURL.BytesValue)
		if urlStr == "" {
			done <- NodeActionResult{Err: fmt.Errorf("URL cannot be empty")}
			return
		}

		// Method (index 1) - Optional, default "GET"
		method := "GET"
		if valMethod, wired, err := n.GetInputValue(1); err == nil && wired {
			m := string(valMethod.BytesValue)
			if m != "" {
				method = strings.ToUpper(m)
			}
		} else if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad Method input: %v", err)}
			return
		}

		// Body (index 2) - Optional
		var bodyReader io.Reader
		if valBody, wired, err := n.GetInputValue(2); err == nil && wired {
			if len(valBody.BytesValue) > 0 {
				bodyReader = bytes.NewReader(valBody.BytesValue)
			}
		} else if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad Body input: %v", err)}
			return
		}

		// Create Request
		// We use context.Background() here because Run() doesn't pass ctx.
		// Use RunContext if available.
		req, err := http.NewRequest(method, urlStr, bodyReader)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to create request: %v", err)}
			return
		}

		// Execute
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("request failed: %v", err)}
			return
		}
		defer func() { _ = resp.Body.Close() }()

		// Read Response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to read response body: %v", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewInt64Value(int64(resp.StatusCode), 0),
				NewBytesValue(respBody),
			},
		}
	}()
	return done
}

func (a *HTTPRequestAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	// Reimplement logic with Context support
	done := make(chan NodeActionResult, 1)
	go func() {
		// Get Inputs
		valURL, _, errURL := n.GetInputValue(0)
		if errURL != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad URL input: %v", errURL)}
			return
		}
		urlStr := string(valURL.BytesValue)
		if urlStr == "" {
			done <- NodeActionResult{Err: fmt.Errorf("URL cannot be empty")}
			return
		}

		method := "GET"
		if valMethod, wired, err := n.GetInputValue(1); err == nil && wired {
			m := string(valMethod.BytesValue)
			if m != "" {
				method = strings.ToUpper(m)
			}
		} else if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad Method input: %v", err)}
			return
		}

		var bodyReader io.Reader
		if valBody, wired, err := n.GetInputValue(2); err == nil && wired {
			if len(valBody.BytesValue) > 0 {
				bodyReader = bytes.NewReader(valBody.BytesValue)
			}
		} else if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad Body input: %v", err)}
			return
		}

		req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to create request: %v", err)}
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("request failed: %v", err)}
			return
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("failed to read response body: %v", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewInt64Value(int64(resp.StatusCode), 0),
				NewBytesValue(respBody),
			},
		}
	}()
	return done
}

func (a *HTTPRequestAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
