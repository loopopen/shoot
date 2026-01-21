package shoot

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"golang.org/x/tools/imports"
)

func MergeSources(sources ...[]byte) ([]byte, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	fset := token.NewFileSet()
	var files []*ast.File

	for i, src := range sources {
		file, err := parser.ParseFile(fset, fmt.Sprintf("src%d.go", i), src, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse src%d.go: %w", i, err)
		}
		files = append(files, file)
	}

	pkgName := files[0].Name.Name

	importSet := MakeSet[string]()
	var importDecls []ast.Decl

	for _, f := range files {
		for _, imp := range f.Imports {
			key := imp.Path.Value
			if imp.Name != nil {
				key += imp.Name.Name
			}
			if !importSet.Has(key) {
				importSet.Adds(key)
				importDecls = append(importDecls, &ast.GenDecl{
					Tok:   token.IMPORT,
					Specs: []ast.Spec{imp},
				})
			}
		}
	}

	var buf bytes.Buffer

	// ---- header ----
	if len(files[0].Comments) > 0 {
		fmt.Fprint(&buf, "// ")
		fmt.Fprintln(&buf, files[0].Comments[0].Text())
		fmt.Fprintln(&buf)
	}

	// ---- package ----
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	// ---- imports ----
	if len(importDecls) > 0 {
		for _, decl := range importDecls {
			printer.Fprint(&buf, fset, decl)
			fmt.Fprintln(&buf)
		}
		fmt.Fprintln(&buf)
	}

	// ---- decls ----
	for _, f := range files {
		for _, decl := range f.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}

			if err := printDeclWithOwnComments(&buf, fset, f, decl, pkgName); err != nil {
				return nil, fmt.Errorf("print decl: %w", err)
			}
			fmt.Fprintln(&buf)
		}
	}

	out, err := FormatSrc(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format merged source: %w", err)
	}
	return noopFix(out), nil
}

func printDeclWithOwnComments(buf *bytes.Buffer, fset *token.FileSet, file *ast.File, decl ast.Decl, pkgName string) error {
	var bf bytes.Buffer
	mf := &ast.File{
		Name:  file.Name,
		Decls: []ast.Decl{decl},
	}

	mf.Comments = append(mf.Comments, attachCommentsForDecl(file, decl)...)

	err := printer.Fprint(&bf, fset, mf)
	if err != nil {
		return err
	}

	reg := regexp.MustCompile("(?m)^package " + pkgName + "$")
	src := bf.Bytes()
	if !reg.Match(src) {
		logx.Fatal("cannot merge fiels with different packages")
	}
	src = reg.ReplaceAll(src, nil)
	buf.Write(src)
	return nil
}

func attachCommentsForDecl(file *ast.File, decl ast.Decl) []*ast.CommentGroup {
	var groups []*ast.CommentGroup
	declStart := decl.Pos()
	declEnd := decl.End()

	for _, cg := range file.Comments {
		if cg.Pos() >= declStart && cg.Pos() <= declEnd {
			groups = append(groups, cg)
			continue
		}

		if cg.End() <= declStart {
			if declStart-cg.End() < 10 {
				groups = append(groups, cg)
			}
		}
	}

	return groups
}

func noopFix(src []byte) []byte {
	noopFuncReg := regexp.MustCompile(`(\w+)\(\)\s*{\s*}`)
	src = noopFuncReg.ReplaceAll(src, []byte("$1() { /*noop*/ }"))
	return src
}

func FormatSrc(src []byte) ([]byte, error) {
	// format imports
	src, err := imports.Process("_.go", src, nil)
	if err != nil {
		return nil, err
	}

	// format source code
	src, err = format.Source(src)
	if err != nil {
		fmt.Println(string(src))
		return nil, err
	}
	return src, nil
}
