package nodes

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type SaveFileAction struct {
	Path   string
	Format string // "raw", "csv", "json"
}

func NewSaveFileNode() *core.Node {
	return &core.Node{
		Name: "Save File",

		InputPorts: []core.NodePort{{
			Name: "Data",
			Type: core.FlowType{Kind: core.FSKindAny},
		}},
		OutputPorts: []core.NodePort{}, // No output, just side effect? Or return path/success?

		Action: &SaveFileAction{
			Path:   "output.txt",
			Format: "raw",
		},
	}
}

var _ core.NodeAction = &SaveFileAction{}

func (c *SaveFileAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.Path)
	core.SStr(s, &c.Format)
	return s.Ok()
}

func (c *SaveFileAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	// Could validate path validity here
}

func (c *SaveFileAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("SaveFileContainer", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:          core.GROWH,
			ChildAlignment:  core.YCENTER,
			LayoutDirection: clay.TopToBottom,
			ChildGap:        core.S2,
		},
	}, func() {
		// Input Port
		clay.CLAY(clay.IDI("SaveFileInputRow", n.ID), clay.EL{Layout: clay.LAY{ChildAlignment: core.YCENTER, Sizing: core.GROWH}}, func() {
			core.UIInputPort(n, 0)
			clay.TEXT("Data", clay.TextElementConfig{TextColor: core.White})
		})

		// Path Input
		clay.CLAY(clay.IDI("SaveFilePathRow", n.ID), clay.EL{Layout: clay.LAY{ChildGap: core.S2, Sizing: core.GROWH}}, func() {
			clay.TEXT("Path:", clay.TextElementConfig{TextColor: core.LightGray})
			core.UITextBox(clay.IDI("SaveFilePath", n.ID), &c.Path, core.UITextBoxConfig{
				El: clay.EL{
					Layout: clay.LAY{Sizing: core.GROWH},
				},
			})
		})

		// Format Selection
		clay.CLAY(clay.IDI("SaveFileFormatRow", n.ID), clay.EL{Layout: clay.LAY{ChildGap: core.S2, Sizing: core.GROWH}}, func() {
			clay.TEXT("Format:", clay.TextElementConfig{TextColor: core.LightGray})
			// Simple dropdown or toggle for now.
			// Let's use a simple cycle button for MVP.
			core.UIButton(clay.IDI("SaveFileFormatBtn", n.ID), core.UIButtonConfig{
				OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
					switch c.Format {
					case "raw":
						c.Format = "csv"
					case "csv":
						c.Format = "json"
					default:
						c.Format = "raw"
					}
				},
			}, func() {
				clay.TEXT(c.Format, clay.TextElementConfig{TextColor: core.White})
			})
		})
	})
}

func (c *SaveFileAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
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
		if !ok {
			res.Err = errors.New("input required")
			return
		}
		if err != nil {
			res.Err = err
			return
		}

		// Open file
		f, err := os.Create(c.Path)
		if err != nil {
			res.Err = fmt.Errorf("failed to create file: %w", err)
			return
		}
		defer func() { _ = f.Close() }()

		switch c.Format {
		case "raw":
			// Expect Bytes or String
			var data []byte
			if input.Type.Kind == core.FSKindBytes {
				data = input.BytesValue
			} else {
				// Try to convert to string representation
				data = []byte(fmt.Sprintf("%v", input)) // Placeholder, should use proper string conversion
				// Better: check specific types
				switch input.Type.Kind {
				case core.FSKindInt64:
					data = []byte(strconv.FormatInt(input.Int64Value, 10))
				case core.FSKindFloat64:
					data = []byte(strconv.FormatFloat(input.Float64Value, 'f', -1, 64))
				}
			}
			_, err = f.Write(data)
			if err != nil {
				res.Err = err
				return
			}

		case "json":
			// Marshal whatever we have
			// Need a way to convert core.FlowValue to Go types that json.Marshal understands?
			// Or implement Marshaler on core.FlowValue?
			// For now, let's just use a simple recursive converter or just dump core.FlowValue structure (which might be too verbose).
			// Ideally we want the "value" not the core.FlowValue struct.
			// Let's make a helper toToNative(core.FlowValue) interface{}

			// Quick implementation of ToNative
			native := core.FlowValueToNative(input)
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			err = enc.Encode(native)
			if err != nil {
				res.Err = err
				return
			}

		case "csv":
			if input.Type.Kind != core.FSKindTable {
				res.Err = errors.New("CSV format requires Table input")
				return
			}

			w := csv.NewWriter(f)
			defer w.Flush()

			// Write header
			var header []string
			if input.Type.ContainedType != nil {
				for _, field := range input.Type.ContainedType.Fields {
					header = append(header, field.Name)
				}
			}
			if err := w.Write(header); err != nil {
				res.Err = err
				return
			}

			// Write rows
			for _, row := range input.TableValue {
				select {
				case <-ctx.Done():
					res.Err = ctx.Err()
					return
				default:
				}
				var record []string
				for _, field := range row {
					// Convert value to string
					valStr := fmt.Sprintf("%v", field.Value) // Simplification
					// Proper conversion:
					v := field.Value
					switch v.Type.Kind {
					case core.FSKindBytes:
						valStr = string(v.BytesValue)
					case core.FSKindInt64:
						valStr = strconv.FormatInt(v.Int64Value, 10)
					case core.FSKindFloat64:
						valStr = strconv.FormatFloat(v.Float64Value, 'f', -1, 64)
					}
					record = append(record, valStr)
				}
				if err := w.Write(record); err != nil {
					res.Err = err
					return
				}
			}

		default:
			res.Err = fmt.Errorf("unknown format: %s", c.Format)
			return
		}

		// Success
		res.Outputs = []core.FlowValue{} // No outputs
	}()

	return done
}

func (c *SaveFileAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return c.RunContext(context.Background(), n)
}
