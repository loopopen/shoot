package constructor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"golang.org/x/tools/go/packages"
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
		if g.flags.getset {
			for _, decl := range f.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if ts.Name.Name != typeName {
						continue
					}

					if gd.Doc != nil {
						getter, setter := parseGetterSetter(gd)
						g.getter = getter
						g.setter = setter
					}
				}
			}
		}

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

func (g *Generator) extractTopFiels(pkg *packages.Package, st *ast.StructType, fields *[]*Field) {
	for _, f := range st.Fields.List {
		// if f.Tag != nil {
		// }

		isNew := parseNewComment(f.Doc.Text())
		if isNew {
			g.hasNew = isNew
		}

		if len(f.Names) == 0 {
			//embedded: gorm.Model
			typ := pkg.TypesInfo.TypeOf(f.Type)
			expandIfStruct(pkg, g.qualifier, 0, typ, isNew, fields)
			continue
		}

		//named:
		for _, name := range f.Names {
			obj, ok := pkg.TypesInfo.Defs[name].(*types.Var)
			if !ok {
				continue
			}

			if strings.HasPrefix(name.Name, "_") {
				continue
			}

			if f.Tag != nil {
				new := parseNewTag(f.Tag.Value)
				if new == "-" {
					continue
				}
			}

			var get, set bool
			if g.flags.getset {
				get, set = parseGetSet(f, name.Name)
			}
			defv := parseDef(f)
			var tag string
			if g.flags.json && f.Tag != nil {
				tag = parseJSONTag(f.Tag.Value)
			}

			qname, isPtr := qualifiedName(obj.Type(), g.qualifier)
			checkShadowAndAppend(fields, &Field{
				name:          name.Name,
				qualifiedType: qname,
				typ:           obj.Type(),
				depth:         0,
				isPtr:         isPtr,
				isGet:         get,
				isSet:         set,
				isNew:         isNew,
				defValue:      defv,
				jsonTag:       tag,
			})
		}
	}
}

func parseGetSet(f *ast.Field, name string) (bool, bool) {
	var get, set bool
	var g, s bool
	if f.Doc != nil {
		g, s = parseGetSetComment(f.Doc.Text())
		if g == s {
			get = true
			set = true
		} else {
			get = g
			set = s
		}
	} else {
		get = true
		set = true
	}
	if ast.IsExported(name) {
		if g || s {
			logx.Fatalf("exported field %s should not has get/set flag", name)
		}
		return false, false
	}
	return get, set
}

func parseDef(f *ast.Field) string {
	if f.Doc != nil {
		v, ok := parseDefComment(f.Doc.Text())
		if ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func newParamsList(fields []*Field, nameMap map[string]string) string {
	var buf bytes.Buffer
	for _, f := range fields {
		if f.isEmbeded || f.isShadowed {
			continue
		}
		name, ok := nameMap[f.name]
		if !ok {
			continue
		}
		buf.WriteString(name)
		buf.WriteString(" ")
		if f.isPtr {
			buf.WriteString("*")
		}
		buf.WriteString(f.qualifiedType)
		buf.WriteString(", ")
	}
	lst := buf.String()
	if len(lst) > 2 {
		lst = lst[:len(lst)-2]
	}
	return lst
}

func newBody(fields []*Field, nameMap map[string]string) string {
	if len(fields) == 0 {
		return ""
	}
	var buf bytes.Buffer
	newBodyRec(&buf, fields, 0, -1, nameMap)
	return buf.String()
}

func newBodyRec(buf *bytes.Buffer, fields []*Field, pointer int, depth int32, nameMap map[string]string) int {
	if len(fields) == 0 {
		return 0
	}
	var i = pointer
	for i < len(fields) {
		f := fields[i]
		if f.depth <= depth {
			break
		}
		if f.isEmbeded {
			ref := ""
			if f.isPtr {
				ref = "&"
			}
			buf.WriteString(fmt.Sprintf("%s: %s%s{\n", f.name, ref, f.qualifiedType))
			i = newBodyRec(buf, fields, i+1, f.depth, nameMap)
			buf.WriteString("},\n")
		} else {
			name, ok := nameMap[f.name]
			if ok && !f.isShadowed {
				buf.WriteString(fmt.Sprintf("%s: %s,\n", f.name, name))
			} else if f.defValue != "" {
				buf.WriteString(fmt.Sprintf("%s: %s,\n", f.name, f.defValue))
			}
			i++
		}
	}
	return i
}

func expandIfStruct(pkg *packages.Package, qf types.Qualifier, depth int32, t types.Type, isNew bool, fields *[]*Field) {
	var stru *types.Struct
	switch tt := t.(type) {
	case *types.Pointer:
		e := tt.Elem()
		if st, ok := e.Underlying().(*types.Struct); ok {
			stru = st
			//todo: embeded struct?
		}
	case *types.Named:
		if st, ok := tt.Underlying().(*types.Struct); ok {
			stru = st
		}
	case *types.Struct: //todo: embeded struct?
		stru = tt
	}
	if stru != nil {
		qname, isPtr := qualifiedName(t, qf)
		checkShadowAndAppend(fields, &Field{
			name:          shortName(t),
			qualifiedType: qname,
			depth:         depth,
			isPtr:         isPtr,
			isEmbeded:     true,
			typ:           t,
		})
		extractStructFields(pkg, qf, depth+1, stru, isNew, fields)
	}
}

func extractStructFields(pkg *packages.Package, qf types.Qualifier, depth int32, st *types.Struct, isNew bool, fields *[]*Field) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)

		if f.Embedded() {
			expandIfStruct(pkg, qf, depth, f.Type(), isNew, fields)
			continue
		}

		checkShadowAndAppend(fields, &Field{
			name:          f.Name(),
			qualifiedType: types.TypeString(f.Type(), qf),
			depth:         depth,
			isNew:         isNew,
			typ:           f.Type(),
		})
	}
}

func checkShadowAndAppend(fields *[]*Field, field *Field) {
	for _, f := range *fields {
		if f.name != field.name {
			continue
		}
		if field.depth < f.depth {
			f.isShadowed = true
		} else if field.depth > f.depth {
			field.isShadowed = true
		}
	}
	*fields = append(*fields, field)
}

func shortName(t types.Type) string {
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

func qualifiedName(t types.Type, qf types.Qualifier) (string, bool) {
	name := types.TypeString(t, qf)
	isPtr := false
	if strings.HasPrefix(name, "*") {
		isPtr = true
		name = strings.TrimLeft(name, "*")
	}
	return name, isPtr
}

func parseGetSetComment(doc string) (bool, bool) {
	regGet := regexp.MustCompile(`(?im)^shoot:.*?\Wget(;.*|\s*)$`)
	regSet := regexp.MustCompile(`(?im)^shoot:.*?\Wset(;.*|\s*)$`)
	get := regGet.MatchString(doc)
	set := regSet.MatchString(doc)
	return get, set
}

func parseNewComment(doc string) bool {
	regNew := regexp.MustCompile(`(?im)^shoot:.*?\Wnew(;.*|\s*)$`)
	new := regNew.MatchString(doc)
	return new
}

func parseDefComment(doc string) (string, bool) {
	regDef := regexp.MustCompile(`(?im)^shoot:.*?\Wdef(ault)?=([^;\n]+)(;.*|\s*)$`)
	ms := regDef.FindStringSubmatch(doc)
	for idx, m := range ms {
		if (m == "" || m == "ault") && idx+1 < len(ms) {
			return ms[idx+1], true
		}
	}
	return "", false
}

func parseJSONTag(tag string) string {
	reg := regexp.MustCompile(`json:"([^"]*)"`)
	matches := reg.FindStringSubmatch(tag)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}

func parseNewTag(tag string) string {
	reg := regexp.MustCompile(`new:"([^"]*)"`)
	matches := reg.FindStringSubmatch(tag)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}

func parseGetterSetter(genDecl *ast.GenDecl) (bool, bool) {
	var getter, setter bool
	if genDecl.Doc != nil {
		g, s := parseGetterSetterDoc(genDecl.Doc.Text())
		if g == s {
			getter = true
			setter = true
		} else {
			getter = g
			setter = s
		}
	} else {
		return true, true
	}
	return getter, setter
}

func parseGetterSetterDoc(doc string) (bool, bool) {
	regGet := regexp.MustCompile(`(?im)^shoot:.*?\Wgetter(;.*|\s*)$`)
	regSet := regexp.MustCompile(`(?im)^shoot:.*?\Wsetter(;.*|\s*)$`)
	get := regGet.MatchString(doc)
	set := regSet.MatchString(doc)
	return get, set
}
