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

	importSet := map[string]bool{}
	var importDecls []ast.Decl
	for _, f := range files {
		for _, imp := range f.Imports {
			key := imp.Path.Value
			if imp.Name != nil {
				key = key + imp.Name.Name
			}
			if !importSet[key] {
				importSet[key] = true
				importDecls = append(importDecls, &ast.GenDecl{
					Tok:   token.IMPORT,
					Specs: []ast.Spec{imp},
				})
			}
		}
	}

	var otherDecls []ast.Decl
	for _, f := range files {
		for _, decl := range f.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}
			otherDecls = append(otherDecls, decl)
		}
	}

	var buf bytes.Buffer
	// header (first line comment)
	if len(files[0].Comments) > 0 {
		fmt.Fprint(&buf, "// ")
		fmt.Fprintln(&buf, files[0].Comments[0].Text())
		fmt.Fprintln(&buf)
	}
	// package
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)
	// imports
	if len(importDecls) > 0 {
		for _, decl := range importDecls {
			printer.Fprint(&buf, fset, decl)
			fmt.Fprintln(&buf)
		}
		fmt.Fprintln(&buf)
	}
	// decls
	for _, decl := range otherDecls {
		printer.Fprint(&buf, fset, decl)
		fmt.Fprintln(&buf)
	}

	out, err := FormatSrc(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format merged source: %w", err)
	}
	return noopFix(out), nil
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
