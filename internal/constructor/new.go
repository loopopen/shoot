package constructor

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"regexp"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) makeNew(typeName string) {
	g.RegisterTransfer("typeof", transfer.ID)

	var imports string
	var allList []string
	var unexportedList []string
	var newList []string
	var embedList []string
	typeMap := make(map[string]string)
	xMap := make(map[string]string)

	pkgPath := g.Pkg().PkgPath
	if pkgPath == shoot.SelfPkgPath {
		g.data.Self = true
	}

	typeExists := false
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)

			typeExists = true
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					xMap = parseEmbedField(g.Pkg(), field)
					for typ := range xMap {
						embedList = append(embedList, typ)
					}
				}

				for _, name := range field.Names {
					if name.Obj.Kind != ast.Var {
						continue
					}

					fs := token.NewFileSet()
					typeMap[name.Name] = exprString(fs, field.Type)

					allList = append(allList, name.Name)

					if ast.IsExported(name.Name) {
						continue
					}

					unexportedList = append(unexportedList, name.Name)
					if field.Doc != nil {
						n := parseNew(field.Doc.Text())
						if n {
							newList = append(newList, name.Name)
						}
					}
				}
			}

			imports = buildImports(f.Imports)
			return false
		})
	}
	if !typeExists {
		logx.Fatalf("type not exists: %s", typeName)
	}

	g.data.Imports = imports
	g.data.AllList = allList
	if len(newList) > 0 {
		g.data.NewList = newList
	} else if g.flags.exp {
		g.data.NewList = allList
	} else {
		g.data.NewList = unexportedList
	}
	g.data.EmbedList = embedList

	g.RegisterTransfer("typeof", func(key string) string {
		return typeMap[key]
	})
}

func buildImports(imports []*ast.ImportSpec) string {
	var buff bytes.Buffer
	for _, imp := range imports {
		if imp.Name != nil {
			buff.WriteString(imp.Name.Name)
			buff.WriteString(" ")
		}
		if imp.Path != nil {
			buff.WriteString(imp.Path.Value)
			buff.WriteString("\n")
		}
	}
	return buff.String()
}

func parseNew(doc string) bool {
	regNew := regexp.MustCompile(`(?im)^shoot:.*?\Wnew(;.*|\s*)$`)
	new := regNew.MatchString(doc)
	return new
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, expr)
	if err != nil {
		logx.Fatalf("print expr: %s", err)
	}
	return buf.String()
}

func parseEmbedField(pkg *packages.Package, field *ast.Field) map[string]string {
	//todo: only supports 1 depth, recursively ref shoot map

	if !hasFields(pkg.TypesInfo.TypeOf(field.Type)) {
		return nil
	}

	selMap := make(map[string]string)
	switch t := field.Type.(type) {
	case *ast.Ident:
		selMap[t.Name] = ""
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			selMap[t.Sel.Name] = pkgIdent.Name
		}
	case *ast.StarExpr:
		switch x := t.X.(type) {
		case *ast.Ident:
			selMap[x.Name] = ""
		case *ast.SelectorExpr:
			if pkgIdent, ok := x.X.(*ast.Ident); ok {
				selMap[x.Sel.Name] = pkgIdent.Name
			}
		}
	}

	return selMap
}

func hasFields(t types.Type) bool {
	var st *types.Struct
	switch tt := t.(type) {
	case *types.Pointer:
		e := tt.Elem()
		st, _ = e.Underlying().(*types.Struct)
	case *types.Named:
		st, _ = tt.Underlying().(*types.Struct)
	case *types.Struct:
		st = tt
	}
	if st != nil {
		return st.NumFields() > 0
	}
	return false
}
