package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsageRecordInputsCarryProductSettlement(t *testing.T) {
	files := []string{
		"gateway_handler.go",
		"gateway_handler_responses.go",
		"gateway_handler_chat_completions.go",
		"gemini_v1beta_handler.go",
		"openai_chat_completions.go",
		"openai_gateway_handler.go",
		"openai_images.go",
		"sora_gateway_handler.go",
	}

	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
			require.NoError(t, err)

			checked := 0
			ast.Inspect(file, func(node ast.Node) bool {
				unary, ok := node.(*ast.UnaryExpr)
				if !ok || unary.Op != token.AND {
					return true
				}
				lit, ok := unary.X.(*ast.CompositeLit)
				if !ok || !isUsageRecordInputType(lit.Type) {
					return true
				}
				checked++
				require.Truef(t, hasCompositeField(lit, "ProductSettlement"), "%s:%d usage record input is missing ProductSettlement", name, fset.Position(lit.Pos()).Line)
				return true
			})
			require.NotZero(t, checked, "expected to find at least one usage record input")
		})
	}
}

func isUsageRecordInputType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "service" {
		return false
	}
	switch selector.Sel.Name {
	case "RecordUsageInput", "RecordUsageLongContextInput", "OpenAIRecordUsageInput":
		return true
	default:
		return false
	}
}

func hasCompositeField(lit *ast.CompositeLit, name string) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if ok && key.Name == name {
			return true
		}
	}
	return false
}
