package constructor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) parseFields(typeName string) {
	var imports string

	var typeParams []string
	typeParamsMap := make(map[int]string)

	pkgPath := g.Pkg().PkgPath
	if pkgPath == shoot.SelfPkgPath {
		g.data.Self = true
	}

	var fields []*Field

	typeExists := false
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			typeExists = true
			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)

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
	if !typeExists {
		logx.Fatalf("type not exists: %s", typeName)
	}

	g.data.Imports = imports

	g.typeParams = typeParams
	g.typeParamsMap = typeParamsMap
	g.fields = fields
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
	var embedList []string
	nameMap := make(map[string]string)
	typeMap := make(map[string]string)
	var defList []string
	defValueMap := make(map[string]string)
	var getterIfaces []string
	var setterIfaces []string
	var expList []string
	for _, f := range g.fields {
		if f.isShadowed {
			continue
		}
		if f.isEmbeded {
			embedList = append(embedList, f.name)
			typ := f.typ
			if !f.isPtr {
				typ = types.NewPointer(typ)
			}
			get, set := g.findGetterSetterIfac(f.name)
			if get != nil && types.ConvertibleTo(typ, get) {
				getterIfaces = append(getterIfaces, types.TypeString(get, g.qualifier))
			}
			if set != nil && types.ConvertibleTo(typ, set) {
				setterIfaces = append(setterIfaces, types.TypeString(set, g.qualifier))
			}
			continue
		}

		allList = append(allList, f.name)

		if ast.IsExported(f.name) {
			expList = append(expList, f.name)
		}

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
	g.data.EmbedList = embedList
	g.data.GetterIfaces = getterIfaces
	g.data.SetterIfaces = setterIfaces
	g.data.NewMap = nameMap
	g.data.DefaultList = defList
	g.data.DefaultValueMap = defValueMap
	g.data.ExportedList = expList
	g.data.Option = g.flags.opt
	g.data.Short = g.flags.short
}

func (g *Generator) findGetterSetterIfac(name string) (types.Type, types.Type) {
	getterName := name + "Getter"
	setterName := name + "Setter"
	var get, set types.Type
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(node ast.Node) bool {
			ts, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, ok := ts.Type.(*ast.InterfaceType); !ok {
				return true
			}

			if ts.Name.Name != getterName && ts.Name.Name != setterName {
				return true
			}

			obj := g.Pkg().TypesInfo.Defs[ts.Name]
			if obj == nil {
				return true
			}

			named, ok := obj.Type().(*types.Named)
			if !ok {
				return true
			}

			switch ts.Name.Name {
			case getterName:
				get = named
			case setterName:
				set = named
			}
			return get != nil && set != nil
		})
	}
	return get, set
}
