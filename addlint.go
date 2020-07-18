package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	analyzer := &analysis.Analyzer{
		Name: "addlint",
		Doc:  "addlint",
		Run:  run,
	}
	singlechecker.Main(analyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			be, ok := n.(*ast.BinaryExpr)
			if !ok {
				return true
			}

			if be.Op != token.ADD {
				return true
			}

			if _, ok := be.X.(*ast.BasicLit); !ok {
				return true
			}

			if _, ok := be.Y.(*ast.BasicLit); !ok {
				return true
			}
			left := int64(0)
			right := int64(0)

			parseInt := func(expr ast.Expr, dest *int64) bool {
				var err error
				t := pass.TypesInfo.TypeOf(expr)
				if t == nil {
					return false
				}

				bt, ok := t.Underlying().(*types.Basic)
				if !ok {
					return false
				}

				if (bt.Info() & types.IsInteger) == 0 {
					return false
				}
				leftExpr, ok := expr.(*ast.BasicLit)
				if !ok {
					return false
				}
				*dest, err = strconv.ParseInt(leftExpr.Value, 10, 64)
				if err != nil {
					return false
				}
				return true
			}

			if !parseInt(be.X, &left) || !parseInt(be.Y, &right) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: be.Pos(), Message: "fix",
				SuggestedFixes: []analysis.SuggestedFix{
					{Message: "fix", TextEdits: []analysis.TextEdit{
						{Pos: be.Pos(), End: be.End(), NewText: []byte(fmt.Sprintf("%d + %d", left, right))},
					}},
				},
			})

			return true
		})
	}

	return nil, nil
}

func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
