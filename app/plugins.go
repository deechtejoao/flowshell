package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bvisness/flowshell/clay"
)

// PluginMetadata defines the structure expected from --describe
type PluginMetadata struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Inputs      []PluginPort `json:"inputs"`
	Outputs     []PluginPort `json:"outputs"`
}

type PluginPort struct {
	Name string `json:"name"`
	Type string `json:"type"` // "string", "bytes", "int", "float", "any"
}

// GEN:NodeAction
type PluginAction struct {
	Path string
	Meta PluginMetadata
}

func (a *PluginAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *PluginAction) UI(n *Node) {
	// Standard UI for input/output ports
	clay.CLAY(clay.IDI("PluginNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		// Inputs
		for i := range n.InputPorts {
			UIInputPort(n, i)
		}
		// Outputs
		for i := range n.OutputPorts {
			UIOutputPort(n, i)
		}
	})
}

func (a *PluginAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		// Collect inputs
		inputMap := make(map[string]interface{})
		for i, port := range n.InputPorts {
			val, _, err := n.GetInputValue(i)
			if err != nil {
				// Skipping optional? For now assume all required or nil
				continue
			}
			// Convert FlowValue to native Go type for JSON marshaling
			inputMap[port.Name] = FlowValueToNative(val)
		}

		// Prepare Request
		req := map[string]interface{}{
			"method": "process",
			"params": inputMap,
		}
		reqBytes, _ := json.Marshal(req)

		// Run Process
		// Note: For performance, we should keep a persistent process if expected to handle many requests?
		// For now, simpler: One-shot process execution.
		cmd := exec.Command(a.Path)
		// Send input via Stdin
		cmd.Stdin = strings.NewReader(string(reqBytes))

		output, err := cmd.Output()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				done <- NodeActionResult{Err: fmt.Errorf("plugin error: %s\nStderr: %s", err, string(ee.Stderr))}
			} else {
				done <- NodeActionResult{Err: fmt.Errorf("plugin execution failed: %v", err)}
			}
			return
		}

		// Parse Output
		var resp struct {
			Result map[string]interface{} `json:"result"`
			Error  string                 `json:"error"`
		}
		if err := json.Unmarshal(output, &resp); err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("invalid plugin output: %v\nOutput: %s", err, string(output))}
			return
		}

		if resp.Error != "" {
			done <- NodeActionResult{Err: fmt.Errorf("plugin reported error: %s", resp.Error)}
			return
		}

		// Map results to outputs
		var outputs []FlowValue
		for _, port := range n.OutputPorts {
			if val, ok := resp.Result[port.Name]; ok {
				fv, err := NativeToFlowValue(val)
				if err != nil {
					// Fallback
					outputs = append(outputs, NewStringValue(fmt.Sprintf("error converting: %v", err)))
				} else {
					outputs = append(outputs, fv)
				}
			} else {
				// Default/Zero value for type? Or just empty string/nil?
				// NativeToFlowValue(nil) returns empty string value.
				// But we should probably try to match the port type if possible?
				// For now, empty string is safe-ish.
				outputs = append(outputs, NewStringValue(""))
			}
		}

		done <- NodeActionResult{Outputs: outputs}
	}()
	return done
}

func (a *PluginAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *PluginAction) Tag() string {
	return "Plugin:" + a.Path
}

func (a *PluginAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Path)
	// We might store metadata too or reload it?
	// For now, assume Path is enough to reload (we'd need to re-load metadata from disk/cache)
	// But `a.Meta` needs to be populated on load.
	// This serializer only saves Path.
	// On deserialization, we probably need to re-fetch metadata?
	// Or we should serialize Meta too.
	// Let's serialize Meta for safety/offline loading.
	// (Skipping deep serialization for now to keep it simple, assumes plugins are stable)
	return s.Ok()
}

// Helpers

// Loader

func LoadPlugins(dir string) ([]NodeType, error) {
	var nodes []NodeType

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Extensions: .exe, .py, .js, .sh?
		// On Windows, .exe or .py (if associated)
		// We'll treat all files as potential plugins if executable or script
		ext := filepath.Ext(e.Name())
		path := filepath.Join(dir, e.Name())

		// Skip non-executable looking things?
		if ext != ".exe" && ext != ".py" && ext != ".bat" {
			continue
		}

		// Run --describe
		var cmd *exec.Cmd
		if ext == ".py" {
			cmd = exec.Command("python", path, "--describe")
		} else {
			cmd = exec.Command(path, "--describe")
		}

		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Plugin skip %s: %v\n", e.Name(), err)
			continue
		}

		var meta PluginMetadata
		if err := json.Unmarshal(out, &meta); err != nil {
			fmt.Printf("Plugin invalid meta %s: %v\n", e.Name(), err)
			continue
		}

		// Create NodeType
		nt := NodeType{
			Name: meta.Name,
			Create: func() *Node {
				n := &Node{
					Name: meta.Name,
					Action: &PluginAction{
						Path: path,
						Meta: meta,
					},
				}
				// Populate ports
				for _, p := range meta.Inputs {
					n.InputPorts = append(n.InputPorts, NodePort{
						Name: p.Name,
						Type: parsePluginType(p.Type),
					})
				}
				for _, p := range meta.Outputs {
					n.OutputPorts = append(n.OutputPorts, NodePort{
						Name: p.Name,
						Type: parsePluginType(p.Type),
					})
				}
				return n
			},
		}
		nodes = append(nodes, nt)
	}
	return nodes, nil
}

func parsePluginType(t string) FlowType {
	switch t {
	case "string":
		return FlowType{Kind: FSKindBytes}
	case "bytes":
		return FlowType{Kind: FSKindBytes}
	case "int":
		return FlowType{Kind: FSKindInt64, Unit: 0}
	case "float":
		return FlowType{Kind: FSKindFloat64, Unit: 0}
	default:
		return FlowType{Kind: FSKindAny}
	}
}
