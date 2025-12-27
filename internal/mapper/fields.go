package mapper

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) parseSrcFields(srcTypeName string) types.Type {
	var typ types.Type
	g.tagMap = make(map[string]string)

	ptrTypeMap := make(map[string]string)
	var fields []Field
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(srcTypeName, n) {
				return true
			}

			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			obj := g.Pkg().TypesInfo.Defs[ts.Name]
			if obj != nil {
				typ = obj.Type()
			}

			g.extractExportedTopFiels(g.Pkg(), st, ptrTypeMap, &fields)
			return false
		})
	}
	if typ == nil {
		return nil
	}
	g.srcPtrTypeMap = ptrTypeMap
	for _, f := range fields {
		if ast.IsExported(f.name) {
			g.exportedFields = append(g.exportedFields, f)
			g.data.SrcFieldList = append(g.data.SrcFieldList, f.name)
		} else {
			g.unexportedFields = append(g.unexportedFields, f)
		}
	}
	return typ
}

func (g *Generator) parseDestFields(destTypeName string) types.Type {
	var typ types.Type

	ptrTypeMap := make(map[string]string)
	var exportedFields []Field
	for _, f := range g.destPkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(destTypeName, n) {
				return true
			}

			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			obj := g.destPkg.TypesInfo.Defs[ts.Name]
			if obj != nil {
				typ = obj.Type()
			}

			g.extractExportedTopFiels(g.destPkg, st, ptrTypeMap, &exportedFields)
			return false
		})
	}
	if typ == nil {
		return nil
	}
	for _, f := range exportedFields {
		if ast.IsExported(f.name) {
			g.destExportedFields = append(g.destExportedFields, f)
		} else {
			g.destUnexportedFields = append(g.destUnexportedFields, f)
		}
	}
	g.destPtrTypeMap = ptrTypeMap
	return typ
}

func getMapTag(tag string) string {
	reg := regexp.MustCompile(`map:"([^"]*)"`)
	matches := reg.FindStringSubmatch(tag)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}

func (g *Generator) extractExportedTopFiels(pkg *packages.Package, st *ast.StructType, ptrTypeMap map[string]string, fields *[]Field) {
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			//embedded: gorm.Model
			typ := pkg.TypesInfo.TypeOf(f.Type)
			name := typeName(typ)
			// if !ast.IsExported(name) {
			// 	continue
			// }
			expandIfStruct(pkg, g.qualifier, name, 1, typ, ptrTypeMap, fields)
			continue
		}
		//named:
		name := f.Names[0].Name
		if f.Tag != nil {
			tag := getMapTag(f.Tag.Value)
			if tag == "-" {
				continue
			}
			if tag != "" {
				name = transfer.ToPascalCase(name)
				tag = transfer.ToPascalCase(tag)
				g.tagMap[name] = tag
			}
		}
		for _, name := range f.Names {
			// if !ast.IsExported(name.Name) {
			// 	continue
			// }

			if obj, ok := pkg.TypesInfo.Defs[name].(*types.Var); ok {
				appendOrReplace(fields, Field{
					name:  name.Name,
					path:  name.Name,
					typ:   obj.Type(),
					depth: 0,
				})
			}
		}
	}
}

func expandIfStruct(pkg *packages.Package, qf types.Qualifier, pre string, depth int32, t types.Type, ptrTypeMap map[string]string, fields *[]Field) {
	switch tt := t.(type) {
	case *types.Pointer:
		e := tt.Elem()
		if st, ok := e.Underlying().(*types.Struct); ok {
			if n, ok := e.(*types.Named); ok {
				ptrTypeMap[pre] = types.TypeString(n, qf)
			}
			//todo: embeded struct?
			extractStructFields(pkg, qf, pre, depth, st, ptrTypeMap, fields)
		}
	case *types.Named:
		if st, ok := tt.Underlying().(*types.Struct); ok {
			extractStructFields(pkg, qf, pre, depth, st, ptrTypeMap, fields)
		}
	case *types.Struct: //todo: embeded struct?
		extractStructFields(pkg, qf, pre, depth, tt, ptrTypeMap, fields)
	}
}

func extractStructFields(pkg *packages.Package, qf types.Qualifier, pre string, depth int32, st *types.Struct, ptrSet map[string]string, fields *[]Field) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		// if !ast.IsExported(f.Name()) {
		// 	continue
		// }

		if f.Embedded() {
			name := typeName(f.Type())
			expandIfStruct(pkg, qf, pre+"."+name, depth+1, f.Type(), ptrSet, fields)
			continue
		}

		appendOrReplace(fields, Field{
			name:  f.Name(),
			path:  pre + "." + f.Name(),
			typ:   f.Type(),
			depth: depth,
		})
	}
}

func appendOrReplace(fields *[]Field, field Field) {
	var f *Field
	for i := range *fields {
		if (*fields)[i].name == field.name {
			f = &(*fields)[i]
			if field.depth < f.depth {
				f.path = field.path
				f.typ = field.typ
				f.depth = field.depth
				break
			}
		}
	}
	if f == nil {
		*fields = append(*fields, field)
	}
}

func typeName(t types.Type) string {
	switch tt := t.(type) {
	case *types.Named:
		return tt.Obj().Name()
	case *types.Pointer:
		if named, ok := tt.Elem().(*types.Named); ok {
			return named.Obj().Name()
		}
	}
	return ""
}
