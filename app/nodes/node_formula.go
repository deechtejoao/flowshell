package nodes

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/clay"
	"github.com/expr-lang/expr"
)

// GEN:NodeAction
type FormulaAction struct {
	Expression string
}

func NewFormulaNode() *core.Node {
	return &core.Node{
		Name: "Formula",
		InputPorts: []core.NodePort{
			{Name: "Input", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Result", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &FormulaAction{Expression: "col(\"A\") + 1"},
	}
}

var _ core.NodeAction = &FormulaAction{}

func (a *FormulaAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *FormulaAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("FormulaUI", n.ID), clay.EL{
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          core.GROWH,
			ChildGap:        core.S2,
		},
	}, func() {
		clay.CLAY(clay.IDI("FormulaRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("FormulaSpacer1", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("FormulaRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH},
		}, func() {
			core.UITextBox(clay.IDI("Expression", n.ID), &a.Expression, core.UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
			})
		})
	})
}

func (a *FormulaAction) Run(n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	go func() {
		// Get Input
		valInput, _, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("bad input: %v", err)}
			return
		}

		// Compile Expression
		_, err = expr.Compile(a.Expression, expr.Env(map[string]any{
			"col": func(name string) any { return nil }, // Dummy environment for checking
		}))
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("bad expression: %v", err)}
			return
		}
		program, err := expr.Compile(a.Expression)
		if err != nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("bad expression: %v", err)}
			return
		}

		// Helper to evaluate single item
		eval := func(item core.FlowValue) (core.FlowValue, error) {
			env := make(map[string]any)

			// Helper to access columns
			env["col"] = func(name string) any {
				// Access field from Record or Table Row
				if item.Type.Kind == core.FSKindRecord {
					// Scan fields
					for i, field := range item.Type.Fields {
						if field.Name == name {
							return core.FlowValueToNative(item.RecordValue[i].Value)
						}
					}
				}
				return nil
			}

			env["Input"] = core.FlowValueToNative(item)

			output, err := expr.Run(program, env)
			if err != nil {
				return core.FlowValue{}, err
			}
			return core.NativeToFlowValue(output)
		}

		// Process
		var result core.FlowValue
		if valInput.Type == nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("input has no type (is the input connected?)")}
			return
		}
		switch valInput.Type.Kind {
		case core.FSKindTable:
			// Iterate rows
			var resList []core.FlowValue
			for _, row := range valInput.TableValue {
				// Construct Record from Row for easier access?
				// Or just pass the row and let helper handle it.
				// A TableValue row is []core.FlowValueField.
				// We need to wrap it into a fake core.FlowValue Record to reuse helper logic?
				// Or just handle []core.FlowValueField directly.

				recVal := core.FlowValue{
					Type: &core.FlowType{
						Kind:   core.FSKindRecord,
						Fields: valInput.Type.ContainedType.Fields,
					},
					RecordValue: row,
				}

				outItem, err := eval(recVal)
				if err != nil {
					done <- core.NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
					return
				}
				resList = append(resList, outItem)
			}
			// Result is List
			// Determine contained type from first item? Or Any?
			resType := core.FlowType{Kind: core.FSKindAny}
			if len(resList) > 0 {
				resType = *resList[0].Type
			}
			result = core.NewListValue(resType, resList)

		case core.FSKindList:
			// Map over list
			var resList []core.FlowValue
			for _, item := range valInput.ListValue {
				outItem, err := eval(item)
				if err != nil {
					done <- core.NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
					return
				}
				resList = append(resList, outItem)
			}
			resType := core.FlowType{Kind: core.FSKindAny}
			if len(resList) > 0 {
				resType = *resList[0].Type
			}
			result = core.NewListValue(resType, resList)

		default:
			// Single Item
			res, err := eval(valInput)
			if err != nil {
				done <- core.NodeActionResult{Err: fmt.Errorf("eval error: %v", err)}
				return
			}
			result = res
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{result},
		}
	}()
	return done
}

func (a *FormulaAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	return a.Run(n)
}

func (a *FormulaAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &a.Expression)
	return s.Ok()
}

// core.FlowValueToNative moved to flowdata.go
