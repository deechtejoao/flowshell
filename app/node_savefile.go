package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type SaveFileAction struct {
	Path   string
	Format string // "raw", "csv", "json"
}

func NewSaveFileNode() *Node {
	return &Node{
		Name: "Save File",

		InputPorts: []NodePort{{
			Name: "Data",
			Type: FlowType{Kind: FSKindAny},
		}},
		OutputPorts: []NodePort{}, // No output, just side effect? Or return path/success?

		Action: &SaveFileAction{
			Path:   "output.txt",
			Format: "raw",
		},
	}
}

var _ NodeAction = &SaveFileAction{}

func (c *SaveFileAction) Serialize(s *Serializer) bool {
	SStr(s, &c.Path)
	SStr(s, &c.Format)
	return s.Ok()
}

func (c *SaveFileAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	// Could validate path validity here
}

func (c *SaveFileAction) UI(n *Node) {
	clay.CLAY(clay.IDI("SaveFileContainer", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:          GROWH,
			ChildAlignment:  YCENTER,
			LayoutDirection: clay.TopToBottom,
			ChildGap:        S2,
		},
	}, func() {
		// Input Port
		clay.CLAY(clay.IDI("SaveFileInputRow", n.ID), clay.EL{Layout: clay.LAY{ChildAlignment: YCENTER, Sizing: GROWH}}, func() {
			UIInputPort(n, 0)
			clay.TEXT("Data", clay.TextElementConfig{TextColor: White})
		})

		// Path Input
		clay.CLAY(clay.IDI("SaveFilePathRow", n.ID), clay.EL{Layout: clay.LAY{ChildGap: S2, Sizing: GROWH}}, func() {
			clay.TEXT("Path:", clay.TextElementConfig{TextColor: LightGray})
			UITextBox(clay.IDI("SaveFilePath", n.ID), &c.Path, UITextBoxConfig{
				El: clay.EL{
					Layout: clay.LAY{Sizing: GROWH},
				},
			})
		})

		// Format Selection
		clay.CLAY(clay.IDI("SaveFileFormatRow", n.ID), clay.EL{Layout: clay.LAY{ChildGap: S2, Sizing: GROWH}}, func() {
			clay.TEXT("Format:", clay.TextElementConfig{TextColor: LightGray})
			// Simple dropdown or toggle for now.
			// Let's use a simple cycle button for MVP.
			UIButton(clay.IDI("SaveFileFormatBtn", n.ID), UIButtonConfig{
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
				clay.TEXT(c.Format, clay.TextElementConfig{TextColor: White})
			})
		})
	})
}

func (c *SaveFileAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
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
			if input.Type.Kind == FSKindBytes {
				data = input.BytesValue
			} else {
				// Try to convert to string representation
				data = []byte(fmt.Sprintf("%v", input)) // Placeholder, should use proper string conversion
				// Better: check specific types
				switch input.Type.Kind {
				case FSKindInt64:
					data = []byte(strconv.FormatInt(input.Int64Value, 10))
				case FSKindFloat64:
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
			// Need a way to convert FlowValue to Go types that json.Marshal understands?
			// Or implement Marshaler on FlowValue?
			// For now, let's just use a simple recursive converter or just dump FlowValue structure (which might be too verbose).
			// Ideally we want the "value" not the FlowValue struct.
			// Let's make a helper toToNative(FlowValue) interface{}

			// Quick implementation of ToNative
			native := FlowValueToNative(input)
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			err = enc.Encode(native)
			if err != nil {
				res.Err = err
				return
			}

		case "csv":
			if input.Type.Kind != FSKindTable {
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
					case FSKindBytes:
						valStr = string(v.BytesValue)
					case FSKindInt64:
						valStr = strconv.FormatInt(v.Int64Value, 10)
					case FSKindFloat64:
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
		res.Outputs = []FlowValue{} // No outputs
	}()

	return done
}

func (c *SaveFileAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}


