// Package exitchecker contains linter to detecxt os.Exit function call from main function of main package.
// It is compatible with golang.org/x/tools/go/analysis/multichecker and can be called like other analizers:
//
//	multichecker.Main(exitchecker.ExitChecker)
package exitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitChecker = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "check for os.Exit calls in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isMainFunc := func(f *ast.FuncDecl) bool {
		return f.Name.Name == "main"
	}

	isOsExit := func(x *ast.SelectorExpr) bool {
		ident, ok := x.X.(*ast.Ident)
		if !ok {
			return false
		}

		if ident.Name == "os" && x.Sel.Name == "Exit" {
			pass.Reportf(ident.NamePos, "os.Exit call in main func of main package")
			return true
		}

		return false
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		main := false
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if isMainFunc(x) {
					main = true
					return true
				}
			case *ast.SelectorExpr:
				if main {
					if isOsExit(x) {
						return false
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
