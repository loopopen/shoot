package mapper

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) parseFields(
	pkg *packages.Package, typeName string,
	tagMap, ptrTypeMap *map[string]string,
	exportedFields, unexportedFields *[]*Field) types.Type {
	var typ types.Type

	*ptrTypeMap = make(map[string]string)
	var fields []*Field
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)

			obj := pkg.TypesInfo.Defs[ts.Name]
			if obj != nil {
				typ = obj.Type()
			}

			var tags map[string]string
			if tagMap != nil {
				tags = *tagMap
			}
			g.extractTopFiels(pkg, st, tags, *ptrTypeMap, &fields)
			return false
		})
	}
	if typ == nil {
		return nil
	}
	for _, f := range fields {
		if ast.IsExported(f.Name) {
			*exportedFields = append(*exportedFields, f)
		} else {
			*unexportedFields = append(*unexportedFields, f)
		}
	}
	return typ
}

func (g *Generator) parseSrcFields(srcTypeName string) types.Type {
	//cleaning is important
	g.exportedFields = nil
	g.unexportedFields = nil
	g.srcTagMap = make(map[string]string)
	g.srcPtrTypeMap = make(map[string]string)
	return g.parseFields(g.Pkg(), srcTypeName, &g.srcTagMap, &g.srcPtrTypeMap, &g.exportedFields, &g.unexportedFields)
}

func (g *Generator) parseDestFields(destTypeName string) types.Type {
	g.destExportedFields = nil
	g.destUnexportedFields = nil
	g.destPtrTypeMap = make(map[string]string)
	return g.parseFields(g.destPkg, destTypeName, nil, &g.destPtrTypeMap, &g.destExportedFields, &g.destUnexportedFields)
}

func getMapTag(tag string) string {
	reg := regexp.MustCompile(`map:"([^"]*)"`)
	matches := reg.FindStringSubmatch(tag)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}

func (g *Generator) extractTopFiels(pkg *packages.Package, st *ast.StructType, tagMap, ptrTypeMap map[string]string, fields *[]*Field) {
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			//embedded: gorm.Model
			typ := pkg.TypesInfo.TypeOf(f.Type)
			name := typeName(typ)
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
			if tag != "" && tagMap != nil {
				name = transfer.ToPascalCase(name)
				tag = transfer.ToPascalCase(tag)
				tagMap[name] = tag
			}
		}
		for _, name := range f.Names {
			if obj, ok := pkg.TypesInfo.Defs[name].(*types.Var); ok {
				appendOrReplace(fields, &Field{
					Name:  name.Name,
					path:  name.Name,
					typ:   obj.Type(),
					depth: 0,
				})
			}
		}
	}
}

func expandIfStruct(pkg *packages.Package, qf types.Qualifier, pre string, depth int32, t types.Type, ptrTypeMap map[string]string, fields *[]*Field) {
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

func extractStructFields(pkg *packages.Package, qf types.Qualifier, pre string, depth int32, st *types.Struct, ptrSet map[string]string, fields *[]*Field) {
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

		appendOrReplace(fields, &Field{
			Name:  f.Name(),
			path:  pre + "." + f.Name(),
			typ:   f.Type(),
			depth: depth,
		})
	}
}

func appendOrReplace(fields *[]*Field, field *Field) {
	var f *Field
	for i := range *fields {
		if (*fields)[i].Name == field.Name {
			f = (*fields)[i]
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
