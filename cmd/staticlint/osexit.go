package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// OsExitCheckAnalyzer проверяет использование os.Exit в функции main пакета main.
//
// Вызов os.Exit в main функции прерывает выполнение программы немедленно,
// не позволяя отложенным функциям (defer) выполниться. Это может привести к:
// - не закрытым файлам и соединениям
// - не записанным логам
// - не выполненным cleanup операциям
//
// Рекомендуется использовать return или log.Fatal вместо прямого вызова os.Exit в main.
var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "check for os.Exit calls in main function of main package",
	Run:      runOsExitCheck,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func runOsExitCheck(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтр для поиска только деклараций функций
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Проверяем только функцию main
		if funcDecl.Name.Name != "main" {
			return
		}

		// Рекурсивно проверяем тело функции на наличие os.Exit
		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			// Ищем вызовы функций
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Проверяем, является ли это вызовом селектора (например, os.Exit)
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Получаем объект, на который ссылается селектор
			obj := pass.TypesInfo.ObjectOf(selExpr.Sel)
			if obj == nil {
				return true
			}

			// Проверяем, что это функция из пакета os с именем Exit
			if pkgName, ok := obj.(*types.Func); ok {
				if pkg := pkgName.Pkg(); pkg != nil && pkg.Name() == "os" && pkgName.Name() == "Exit" {
					pass.Reportf(callExpr.Pos(), "os.Exit should not be called in main function of main package")
				}
			}

			return true
		})
	})

	return nil, nil
}
