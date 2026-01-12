package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type LoadFileAction struct {
	path string

	format UIDropdown

	// TODO: In reality this should be a more complex thing. For now we will just
	// always parse them as floats. (The "right way" to do it would be to have
	// CSV always parse as strings, but make it clear in the UI that they are
	// strings, and then have the user convert them to numbers. Perhaps this
	// could be done with a "Convert to Number" node that works on single
	// strings, lists, records, and tables. But perhaps you'd want to be able to
	// easily apply it to specific columns of a table? Maybe implicit conversion
	// to number would be ok within the Aggregate node and other nodes that do
	// math? Who knows. Very large design space. For now we just demo by always
	// parsing as float.
	inferTypes bool
}

var loadFileFormatOptions = []UIDropdownOption{
	{Name: "Raw bytes", Value: "raw"},
	{Name: "CSV", Value: "csv"},
	{Name: "JSON", Value: "json"},
}

// TODO: Make this node polymorphic on lists of strings
// (rename to "Load Files" dynamically)
func NewLoadFileNode(path string) *Node {
	formatDropdown := UIDropdown{
		Options: loadFileFormatOptions,
	}

	if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
		formatDropdown.SelectByValue(ext[1:])
	}

	return &Node{
		ID:   NewNodeID(),
		Name: "Load File",

		InputPorts: []NodePort{{
			Name: "Path",
			Type: FlowType{Kind: FSKindBytes},
		}},
		OutputPorts: []NodePort{{
			Name: "Data",
			Type: FlowType{Kind: FSKindBytes},
		}},

		Action: &LoadFileAction{
			path:       path,
			format:     formatDropdown,
			inferTypes: true,
		},
	}
}

var _ NodeAction = &LoadFileAction{}

func (c *LoadFileAction) UpdateAndValidate(n *Node) {
	switch c.format.GetSelectedOption().Value {
	case "raw":
		n.OutputPorts[0].Type = FlowType{Kind: FSKindBytes}
	case "csv":
		n.OutputPorts[0].Type = FlowType{Kind: FSKindTable, ContainedType: &FlowType{Kind: FSKindAny}}
	case "json":
		n.OutputPorts[0].Type = FlowType{Kind: FSKindAny}
	}

	n.Valid = true
}

func (c *LoadFileAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{
				Sizing:         GROWH,
				ChildAlignment: YCENTER,
			},
		}, func() {
			PortAnchor(n, false, 0)
			UITextBox(clay.IDI("LoadFilePath", n.ID), &c.path, UITextBoxConfig{
				El: clay.EL{
					Layout: clay.LAY{Sizing: GROWH},
				},
				Disabled: n.InputIsWired(0),
			})
			UISpacer(clay.AUTO_ID, W2)
			UIOutputPort(n, 0)
		})

		c.format.Do(clay.AUTO_ID, UIDropdownConfig{
			El: clay.EL{
				Layout: clay.LAY{Sizing: GROWH},
			},
			OnChange: func(before, after any) {
				n.ClearResult()
			},
		})

		if c.format.GetSelectedOption().Value == "csv" {
			UICheckbox(clay.IDI("InferTypes", n.ID), &c.inferTypes, "Infer Types")
		}
	})
}

func (c *LoadFileAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		var res NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		var path string
		wirePath, hasWire, err := n.GetInputValue(0)
		if err != nil {
			res.Err = err
			return
		}
		if hasWire && wirePath.Type.Kind == FSKindBytes {
			path = string(wirePath.BytesValue)
		} else {
			path = c.path
		}

		f, err := os.Open(path)
		if err != nil {
			res.Err = err
			return
		}
		defer f.Close()

		switch format := c.format.GetSelectedOption().Value; format {
		case "raw":
			content, err := io.ReadAll(f)
			if err != nil {
				res.Err = err
				return
			}
			res = NodeActionResult{
				Outputs: []FlowValue{NewBytesValue(content)},
			}
		case "csv":
			r := csv.NewReader(f)
			r.FieldsPerRecord = -1 // Allow variable number of fields

			// Read all records
			records, err := r.ReadAll()
			if err != nil {
				res.Err = err
				return
			}

			if len(records) == 0 {
				// Special case: if we don't even get a row, synthesize an empty table with no columns.
				res = NodeActionResult{
					Outputs: []FlowValue{{
						Type: &FlowType{
							Kind: FSKindTable,
							ContainedType: &FlowType{
								Kind:   FSKindRecord,
								Fields: nil,
							},
						},
					}},
				}
				return
			}

			header := records[0]
			dataRows := records[1:]
			numCols := len(header)

			// Determine column types
			colTypes := make([]FlowTypeKind, numCols)
			for i := range colTypes {
				colTypes[i] = FSKindBytes // Default to string
			}

			if c.inferTypes {
				for col := 0; col < numCols; col++ {
					isInt := true
					isFloat := true

					for _, row := range dataRows {
						if col >= len(row) {
							continue
						}
						val := row[col]
						if val == "" {
							continue // Empty strings can be anything? Or treated as null/zero? Let's say ignore for type inference.
						}

						if isInt {
							if _, err := strconv.ParseInt(val, 10, 64); err != nil {
								isInt = false
							}
						}
						if isFloat {
							if _, err := strconv.ParseFloat(val, 64); err != nil {
								isFloat = false
							}
						}

						if !isInt && !isFloat {
							break
						}
					}

					if isInt {
						colTypes[col] = FSKindInt64
					} else if isFloat {
						colTypes[col] = FSKindFloat64
					}
				}
			}

			// Build schema
			tableRecordType := FlowType{Kind: FSKindRecord}
			for i, headerField := range header {
				tableRecordType.Fields = append(tableRecordType.Fields, FlowField{
					Name: headerField,
					Type: &FlowType{Kind: colTypes[i]},
				})
			}

			// Build rows
			var tableRows [][]FlowValueField
			for _, row := range dataRows {
				var flowRow []FlowValueField
				for col, value := range row {
					if col >= numCols {
						continue // Ignore extra columns not in header
					}

					var flowValue FlowValue
					switch colTypes[col] {
					case FSKindInt64:
						if value == "" {
							flowValue = NewInt64Value(0, 0) // Handle empty as 0?
						} else {
							val, _ := strconv.ParseInt(value, 10, 64)
							flowValue = NewInt64Value(val, 0)
						}
					case FSKindFloat64:
						if value == "" {
							flowValue = NewFloat64Value(0, 0)
						} else {
							val, _ := strconv.ParseFloat(value, 64)
							flowValue = NewFloat64Value(val, 0)
						}
					default:
						flowValue = NewStringValue(value)
					}

					flowRow = append(flowRow, FlowValueField{
						Name:  header[col],
						Value: flowValue,
					})
				}

				// Handle missing columns (fill with default)
				if len(row) < numCols {
					for col := len(row); col < numCols; col++ {
						var flowValue FlowValue
						switch colTypes[col] {
						case FSKindInt64:
							flowValue = NewInt64Value(0, 0)
						case FSKindFloat64:
							flowValue = NewFloat64Value(0, 0)
						default:
							flowValue = NewStringValue("")
						}
						flowRow = append(flowRow, FlowValueField{
							Name:  header[col],
							Value: flowValue,
						})
					}
				}

				tableRows = append(tableRows, flowRow)
			}

			res = NodeActionResult{
				Outputs: []FlowValue{{
					Type: &FlowType{
						Kind:          FSKindTable,
						ContainedType: &tableRecordType,
					},
					TableValue: tableRows,
				}},
			}
		default:
			res.Err = fmt.Errorf("unknown format \"%v\"", format)
		}
	}()
	return done
}

func (c *LoadFileAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}

func (c *LoadFileAction) Serialize(s *Serializer) bool {
	SStr(s, &c.path)
	SBool(s, &c.inferTypes)

	if s.Encode {
		val := ""
		opt := c.format.GetSelectedOption()
		if v, ok := opt.Value.(string); ok {
			val = v
		}
		s.WriteStr(val)
	} else {
		val, ok := s.ReadStr()
		if !ok {
			return false
		}
		c.format = UIDropdown{Options: loadFileFormatOptions}
		c.format.SelectByValue(val)
	}

	return s.Ok()
}
