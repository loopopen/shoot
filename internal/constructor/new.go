package constructor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) parseFields(typeName string) types.Type {
	var typ types.Type
	var imports string

	var typeParams []string
	typeParamsMap := make(map[int]string)

	pkgPath := g.Pkg().PkgPath
	if pkgPath == shoot.SelfPkgPath {
		g.data.Self = true
	}

	var fields []*Field
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)

			obj := g.Pkg().TypesInfo.Defs[ts.Name]
			if obj != nil {
				typ = obj.Type()
			} else {
				return true
			}

			if ts.TypeParams != nil {
				for i, p := range ts.TypeParams.List {
					var names []string
					for _, n := range p.Names {
						names = append(names, n.Name)
					}
					typeParamsMap[i] = strings.Join(names, ", ")
					if ident, ok := p.Type.(*ast.Ident); ok {
						typeParams = append(typeParams, ident.Name)
					}
				}
			}

			g.extractTopFiels(g.Pkg(), st, &fields)

			imports = buildImports(f.Imports)
			return false
		})
	}
	if typ == nil {
		return nil
	}

	g.typeParams = typeParams
	g.typeParamsMap = typeParamsMap
	g.fields = fields
	g.data.Imports = imports
	return typ
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

func (g *Generator) makeNew() {
	var paramGroups []string
	var nameGroups []string
	for i, t := range g.typeParams {
		names := g.typeParamsMap[i]
		paramGroups = append(paramGroups, fmt.Sprintf("%s %s", names, t))
		nameGroups = append(nameGroups, names)
	}
	g.data.TypeParamList = strings.Join(paramGroups, ", ")
	g.data.TypeParamNameList = strings.Join(nameGroups, ", ")

	var allList []string
	nameMap := make(map[string]string)
	typeMap := make(map[string]string)
	var defList []string
	defValueMap := make(map[string]string)
	for _, f := range g.fields {
		if f.isShadowed {
			continue
		}
		if f.isEmbeded {
			continue
		}

		allList = append(allList, f.name)

		if f.defValue != "" {
			defList = append(defList, f.name)
			defValueMap[f.name] = f.defValue
		}

		star := ""
		if f.isPtr {
			star = "*"
		}
		typeMap[f.name] = star + f.qualifiedType

		if g.hasNew && !f.isNew {
			continue
		}
		nameMap[f.name] = transfer.ToCamelCase(f.name)
	}

	newlst := newParamsList(g.fields, nameMap)
	g.data.NewParamsList = newlst
	body := newBody(g.fields, nameMap)
	g.data.NewBody = body
	g.data.TypeMap = typeMap
	g.data.AllList = allList
	g.data.NewMap = nameMap
	g.data.DefaultList = defList
	g.data.DefaultValueMap = defValueMap
	g.data.Option = g.flags.opt
	g.data.Short = g.flags.short
}
