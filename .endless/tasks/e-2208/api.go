// ── DO NOT EDIT ─────────────────────────────────────────────────────
// This file belongs to E-2208's verify suite and records what was true
// when E-2208 landed. Edit it only if you ARE E-2208. See
// .endless/tasks/CLAUDE.md.
//
// api prints the exported declarations of the Go package in the directory
// named by its argument, without comments or function bodies, one per
// paragraph, sorted, so two trees' output can be diffed for API changes.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, os.Args[1], func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var decls []string
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, d := range f.Decls {
				decls = append(decls, exported(fset, d)...)
			}
		}
	}
	sort.Strings(decls)
	fmt.Println(strings.Join(decls, "\n\n"))
}

func exported(fset *token.FileSet, d ast.Decl) (out []string) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		if !d.Name.IsExported() || (d.Recv != nil && !exportedRecv(d.Recv)) {
			return nil
		}
		d.Doc, d.Body = nil, nil
		d.Recv = unnamed(d.Recv)
		out = append(out, render(fset, d))
	case *ast.GenDecl:
		for _, s := range d.Specs {
			if !exportedSpec(s) {
				continue
			}
			out = append(out, render(fset, &ast.GenDecl{Tok: d.Tok, Specs: []ast.Spec{s}}))
		}
	}
	return out
}

func exportedRecv(recv *ast.FieldList) bool {
	t := recv.List[0].Type
	if s, ok := t.(*ast.StarExpr); ok {
		t = s.X
	}
	if ix, ok := t.(*ast.IndexExpr); ok {
		t = ix.X
	}
	if ix, ok := t.(*ast.IndexListExpr); ok {
		t = ix.X
	}
	id, ok := t.(*ast.Ident)
	return ok && id.IsExported()
}

func exportedSpec(s ast.Spec) bool {
	switch s := s.(type) {
	case *ast.TypeSpec:
		return s.Name.IsExported()
	case *ast.ValueSpec:
		for _, n := range s.Names {
			if n.IsExported() {
				return true
			}
		}
	}
	return false
}

// unnamed drops parameter and result names, which are not part of the API,
// keeping one field per name.
func unnamed(fl *ast.FieldList) *ast.FieldList {
	if fl == nil {
		return nil
	}
	var list []*ast.Field
	for _, f := range fl.List {
		for range max(len(f.Names), 1) {
			list = append(list, &ast.Field{Type: f.Type})
		}
	}
	return &ast.FieldList{List: list}
}

// render prints n without comments, or parameter and result names.
func render(fset *token.FileSet, n ast.Node) string {
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Field:
			n.Doc, n.Comment = nil, nil
		case *ast.TypeSpec:
			n.Doc, n.Comment = nil, nil
		case *ast.ValueSpec:
			n.Doc, n.Comment = nil, nil
		case *ast.FuncType:
			n.Params = unnamed(n.Params)
			n.Results = unnamed(n.Results)
		}
		return true
	})
	var b bytes.Buffer
	_ = printer.Fprint(&b, token.NewFileSet(), n)
	return b.String()
}
