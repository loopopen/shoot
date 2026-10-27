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
	"strconv"
	"strings"

	"github.com/loopopen/shoot/internal/tools/logx"
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

// withoutGeneratedType removes constructor output for typeName from a merged file.
// Other types stay, so files that call their constructors still typecheck.
func withoutGeneratedType(src []byte, typeName string) ([]byte, error) {
	if len(bytes.TrimSpace(src)) == 0 || typeName == "" {
		return src, nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "merged.go", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var decls []ast.Decl
	for _, decl := range file.Decls {
		if generatedForType(decl, typeName) {
			continue
		}
		decls = append(decls, decl)
	}
	file.Decls = decls
	dropUnusedImports(file)
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}

func generatedForType(decl ast.Decl, typeName string) bool {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv != nil {
			return recvBaseName(d.Recv) == typeName
		}
		if d.Name != nil && (d.Name.Name == "New"+typeName || strings.HasSuffix(d.Name.Name, "Of"+typeName)) {
			return true
		}
		return funcOptionForType(d, typeName)
	case *ast.GenDecl:
		if d.Tok != token.TYPE {
			return false
		}
		for _, spec := range d.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil {
				continue
			}
			name := ts.Name.Name
			if name == typeName+"Getter" || name == typeName+"Setter" || name == "_json_"+typeName {
				return true
			}
		}
	}
	return false
}

func funcOptionForType(fn *ast.FuncDecl, typeName string) bool {
	if fn.Type == nil || fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		if optionTypeArg(field.Type) == typeName {
			return true
		}
	}
	return false
}

func optionTypeArg(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.IndexListExpr:
		if len(t.Indices) == 0 {
			return ""
		}
		return typeBaseName(t.Indices[0])
	case *ast.IndexExpr:
		return typeBaseName(t.Index)
	default:
		return ""
	}
}

func recvBaseName(list *ast.FieldList) string {
	if list == nil || len(list.List) == 0 {
		return ""
	}
	return typeBaseName(list.List[0].Type)
}

func typeBaseName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return typeBaseName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return typeBaseName(t.X)
	case *ast.IndexListExpr:
		return typeBaseName(t.X)
	default:
		return ""
	}
}

func dropUnusedImports(file *ast.File) {
	used := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if ok {
			used[id.Name] = true
		}
		return true
	})
	var decls []ast.Decl
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			decls = append(decls, decl)
			continue
		}
		var specs []ast.Spec
		for _, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			if imp.Name != nil && (imp.Name.Name == "_" || imp.Name.Name == ".") {
				specs = append(specs, spec)
				continue
			}
			if used[importLocalName(imp)] {
				specs = append(specs, spec)
			}
		}
		if len(specs) == 0 {
			continue
		}
		gen.Specs = specs
		decls = append(decls, gen)
	}
	file.Decls = decls
}

func importLocalName(imp *ast.ImportSpec) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	path, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		path = imp.Path.Value
	}
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
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
