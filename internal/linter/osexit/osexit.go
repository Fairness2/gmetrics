// Package osexit defines an Analyzer that reports if os.Exit used im main of package
package osexit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "Check os.Exit() is used in main function of package",
	Run:  run,
}

// run анализирует исходные файлы Go на наличие os.Exit() в main функции
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			// Если пакет не main то продолжаем обход
			if file.Name.Name != "main" {
				return true
			}
			// Если функция main то анализируем её
			fn, ok := node.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}
			// Обходим всё дерево функции
			for _, smb := range fn.Body.List {
				ast.Inspect(smb, func(node ast.Node) bool {
					checkNode(node, pass)
					return true
				})
			}

			return true
		})
	}

	return nil, nil
}

// checkNode Проверяем элемент дерева является ли оно вызовом os.Exit()
func checkNode(node ast.Node, pass *analysis.Pass) {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}
	fn, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	fnPack, ok := fn.X.(*ast.Ident)
	if !ok {
		return
	}
	if fn.Sel.Name == "Exit" && fnPack.Name == "os" {
		pass.Reportf(fn.Pos(), "os exit is not prohibited")
	}
}
