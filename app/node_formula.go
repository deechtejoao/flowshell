package app

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/expr-lang/expr"
)

// GEN:NodeAction
type FormulaAction struct {
	Expression string
}

func NewFormulaNode() *Node {
	return &Node{
		Name: "Formula",
		InputPorts: []NodePort{
			{Name: "Input", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &FormulaAction{Expression: "col(\"A\") + 1"},
	}
}

var _ NodeAction = &FormulaAction{}

func (a *FormulaAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *FormulaAction) UI(n *Node) {
	clay.CLAY(clay.IDI("FormulaUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          GROWH,
			ChildGap:        S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("FormulaRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("FormulaSpacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("FormulaRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("Expression", n.ID), &a.Expression, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})
	})
}

func (a *FormulaAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		// Get Input
		valInput, _, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad input: %v", err)}
			return
		}

		// Compile Expression
		_, err = expr.Compile(a.Expression, expr.Env(map[string]any{
			"col": func(name string) any { return nil }, // Dummy environment for checking
		}))
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad expression: %v", err)}
			return
		}
		program, err := expr.Compile(a.Expression)
		if err != nil {
			done <- NodeActionResult{Err: fmt.Errorf("bad expression: %v", err)}
			return
		}

		// Helper to evaluate single item
		eval := func(item FlowValue) (FlowValue, error) {
			env := make(map[string]any)

			// Helper to access columns
			env["col"] = func(name string) any {
				// Access field from Record or Table Row
				if item.Type.Kind == FSKindRecord {
					// Scan fields
					for i, field := range item.Type.Fields {
						if field.Name == name {
							return FlowValueToNative(item.RecordValue[i].Value)
						}
					}
				}
				return nil
			}

			env["Input"] = FlowValueToNative(item)

			output, err := expr.Run(program, env)
			if err != nil {
				return FlowValue{}, err
			}
			return NativeToFlowValue(output)
		}

		// Process
		var result FlowValue
		switch valInput.Type.Kind {
		case FSKindTable:
			// Iterate rows
			var resList []FlowValue
			for _, row := range valInput.TableValue {
				// Construct Record from Row for easier access?
				// Or just pass the row and let helper handle it.
				// A TableValue row is []FlowValueField.
				// We need to wrap it into a fake FlowValue Record to reuse helper logic?
				// Or just handle []FlowValueField directly.

				recVal := FlowValue{
					Type: &FlowType{
						Kind:   FSKindRecord,
						Fields: valInput.Type.ContainedType.Fields,
					},
					RecordValue: row,
				}

				outItem, err := eval(recVal)
				if err != nil {
					done <- NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
					return
				}
				resList = append(resList, outItem)
			}
			// Result is List
			// Determine contained type from first item? Or Any?
			resType := FlowType{Kind: FSKindAny}
			if len(resList) > 0 {
				resType = *resList[0].Type
			}
			result = NewListValue(resType, resList)

		case FSKindList:
			// Map over list
			var resList []FlowValue
			for _, item := range valInput.ListValue {
				outItem, err := eval(item)
				if err != nil {
					done <- NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
					return
				}
				resList = append(resList, outItem)
			}
			resType := FlowType{Kind: FSKindAny}
			if len(resList) > 0 {
				resType = *resList[0].Type
			}
			result = NewListValue(resType, resList)

		default:
			// Single Item
			res, err := eval(valInput)
			if err != nil {
				done <- NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
				return
			}
			result = res
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{result},
		}
	}()
	return done
}

func (a *FormulaAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *FormulaAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Expression)
	return s.Ok()
}

// FlowValueToNative moved to flowdata.go

