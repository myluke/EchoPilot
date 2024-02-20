package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/labstack/gommon/log"
	"github.com/mylukin/EchoPilot/helper"
)

// Generate Bot Events
func GenBotEvents(module string, outFile string) error {
	events := []string{}
	if err := filepath.Walk("./app", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// ignore not .go
		if filepath.Ext(path) != ".go" {
			return nil
		}
		// Don't extract from test files.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		buf, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, buf, parser.AllErrors)
		if err != nil {
			return err
		}

		// file package name
		filePackName := file.Name.Name

		currentPackName := getCurrentPackName(file)
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CallExpr:
				if fn, ok := v.Fun.(*ast.SelectorExpr); ok {
					var packName string
					if pack, ok := fn.X.(*ast.Ident); ok {
						packName = pack.Name
					}
					if packName != currentPackName {
						return true
					}
					funcName := fn.Sel.Name
					if funcName != "SetFSMValue" {
						return true
					}

					var FSMValue *ast.CompositeLit
					if FSMValue, ok = v.Args[1].(*ast.CompositeLit); !ok {
						return false
					}

					for _, elt := range FSMValue.Elts {
						var keyValue *ast.KeyValueExpr
						if keyValue, ok = elt.(*ast.KeyValueExpr); !ok {
							continue
						}
						keyName := keyValue.Key.(*ast.Ident).Name
						if keyName != "NextFn" {
							continue
						}
						var nextFn string
						if se, ok := keyValue.Value.(*ast.SelectorExpr); ok {
							var ppname string
							if pack, ok := se.X.(*ast.Ident); ok {
								ppname = pack.Name
							}
							nextFn = fmt.Sprintf(`%s.%s`, ppname, se.Sel.Name)
						}

						if ident, ok := keyValue.Value.(*ast.Ident); ok {
							nextFn = fmt.Sprintf(`%s.%s`, filePackName, ident.Name)
						}

						if nextFn != "" && !helper.ValueInSlice(nextFn, events) {
							events = append(events, nextFn)
						}

						log.Infof("%s.%s: %s, %v", packName, funcName, keyName, nextFn)
					}
				}
			}
			return true
		})

		return nil
	}); err != nil {
		return err
	}

	var tmpl = template.Must(template.New("i18n").Parse(`// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package routers

import (
	{{if .Data}}app "{{.Package}}/app"{{end}}
	"github.com/labstack/echo/v4"
)

// BotFSMEvents is bot FSM events
var BotFSMEvents = []echo.HandlerFunc{
{{- range $k, $v := .Data }}
	{{$v}},
{{- end }}
}
`))

	goFile, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return tmpl.Execute(goFile, struct {
		Data    []string
		Package string
	}{
		events,
		module,
	})
}

// getCurrentPackName
func getCurrentPackName(file *ast.File) string {
	for _, i := range file.Imports {
		if i.Path.Kind == token.STRING && i.Path.Value == `"github.com/mylukin/EchoPilot"` {
			if i.Name == nil {
				return removeQuotesAndExtractLastPart(i.Path.Value)
			}
			return i.Name.Name
		}
	}
	return ""
}

// removeQuotesAndExtractLastPart 会去除字符串两边的双引号，并返回最后一个斜杠后面的部分。
func removeQuotesAndExtractLastPart(input string) string {
	// 去除两边的双引号
	trimmed := strings.Trim(input, "\"")

	// 分割字符串，获取最后一个部分
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
