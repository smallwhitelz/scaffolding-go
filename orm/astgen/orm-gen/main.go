package main

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"text/template"
)

//go:embed tpl.gohtml
var genOrm string

// 调用这个方法来生成代码
func gen(w io.Writer, srcFile string) error {
	// 语法树解析
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	// 模板渲染
	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	err = tpl.Execute(w, Data{
		File: file,
		Ops: []string{"LT","GT","EQ"},
	})
	if err != nil {
		return err
	}
	return nil
}

type Data struct {
	*File
	Ops []string
}