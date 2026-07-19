// Package linter содержит статический анализатор для проверки использования panic,
// log.Fatal и os.Exit вне функции main пакета main.
package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer — анализатор, сообщающий о любом использовании встроенной функции panic или
// вызовах log.Fatal и os.Exit вне функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name:     "linter",
	Doc:      "проверяет использование panic, log.Fatal и os.Exit",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	inspects := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.CallExpr)(nil),
	}

	var funcStack []string

	inspects.WithStack(nodeFilter, func(n ast.Node, push bool, _ []ast.Node) bool {
		if !push {
			switch n.(type) {
			case *ast.FuncDecl, *ast.FuncLit:
				funcStack = funcStack[:len(funcStack)-1]
			}
			return true
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			funcStack = append(funcStack, node.Name.Name)
		case *ast.FuncLit:
			funcStack = append(funcStack, "<closure>")
		case *ast.CallExpr:
			checkPanic(pass, node)
			checkForbiddenCalls(pass, node, funcStack)
		}
		return true
	})

	return nil, nil
}

func checkPanic(pass *analysis.Pass, node *ast.CallExpr) {
	ident, ok := node.Fun.(*ast.Ident)
	if !ok || ident.Name != "panic" {
		return
	}
	pass.Reportf(node.Pos(), "use of built-in panic is forbidden")
}

func checkForbiddenCalls(pass *analysis.Pass, node *ast.CallExpr, funcStack []string) {
	sel, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	obj := pass.TypesInfo.ObjectOf(sel.Sel)
	if obj == nil {
		return
	}
	pkg := obj.Pkg()
	if pkg == nil {
		return
	}

	if !isForbiddenCall(pkg.Path(), obj.Name()) {
		return
	}

	if pass.Pkg.Name() == "main" && isInsideMainFunc(funcStack) {
		return
	}

	pass.Reportf(node.Pos(), "call to %s.%s is only allowed in the main function of the main package", pkg.Name(), obj.Name())
}

func isForbiddenCall(pkgPath, objName string) bool {
	if pkgPath == "os" && objName == "Exit" {
		return true
	}
	if pkgPath == "log" && objName == "Fatal" {
		return true
	}
	return false
}

func isInsideMainFunc(funcStack []string) bool {
	for _, fName := range funcStack {
		if fName == "main" {
			return true
		}
	}
	return false
}
