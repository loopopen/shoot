package mapper

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"
)

func (g *Generator) loadTypeMapperPkg(typeName string) string {
	//todo: recursive
	var mappers string
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(typeName, n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			st, _ := ts.Type.(*ast.StructType)
			for _, field := range st.Fields.List {
				//mapper must be an embedded field type
				if len(field.Names) > 0 {
					continue
				}

				if field.Tag != nil {
					tag := getMapTag(field.Tag.Value)
					if tag == "-" {
						continue
					}
				}

				//may be ident or selector
				x := field.Type
				if star, ok := field.Type.(*ast.StarExpr); ok {
					x = star.X
				}

				if ident, ok := x.(*ast.Ident); ok {
					if isMapperCandidate(ident, g.Pkg().TypesInfo) {
						g.mapperpkg = g.Pkg()
						mappers = ident.Name
						return false
					}
				}

				sel, ok := x.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				if !isMapperCandidate(sel.Sel, g.Pkg().TypesInfo) {
					continue
				}
				imp := findImportForSelector(f, sel)
				if imp == nil {
					continue
				}
				impPath := strings.Trim(imp.Path.Value, `"`)
				pkgs := g.LoadPackage(impPath)
				g.mapperpkg = pkgs[impPath]
				mappers = sel.Sel.Name
			}
			return false
		})
	}
	return mappers
}

func isMapperCandidate(id *ast.Ident, info *types.Info) bool {
	//a struct with no fields
	obj := info.Uses[id]
	if obj == nil {
		obj = info.Defs[id]
	}
	if obj == nil {
		return false
	}

	t := obj.Type()
	if t == nil {
		return false
	}

	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	return st.NumFields() == 0
}

func (g *Generator) parseMapper(mapperTypeName string) {
	if mapperTypeName == "" {
		return
	}

	var expFuncList []Func
	for _, f := range g.mapperpkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(mapperTypeName, n) {
				return true
			}

			//todo: optimize
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil {
					if len(fn.Recv.List) == 0 {
						continue
					}

					r := fn.Recv.List[0]
					switch expr := r.Type.(type) {
					case *ast.Ident:
						if expr.Name != mapperTypeName {
							continue
						}
					case *ast.StarExpr:
						if ident, ok := expr.X.(*ast.Ident); ok && ident.Name != mapperTypeName {
							continue
						}
					}

					if fn.Type.Params == nil {
						continue
					}
					params := fn.Type.Params.List
					if len(params) != 1 {
						continue
					}
					if fn.Type.Results == nil {
						continue
					}
					results := fn.Type.Results.List
					if len(results) != 1 {
						continue
					}

					expFuncList = append(expFuncList, Func{
						name:   fn.Name.Name,
						param:  g.mapperpkg.TypesInfo.TypeOf(params[0].Type),
						result: g.mapperpkg.TypesInfo.TypeOf(results[0].Type),
					})
				}
			}

			return false
		})
	}
	g.mappingFuncList = expFuncList
}

func (g *Generator) makeMismatch() {
	g.data.SrcToDestFuncMap = make(map[string]string)
	g.data.DestToSrcFuncMap = make(map[string]string)
	g.data.MismatchFuncMap = make(map[string]string)
	g.data.MismatchSubMap = make(map[string]string)
	g.data.DestMismatchSubMap = make(map[string]string)
	g.data.SrcPtrSet = make(map[string]bool)
	g.data.DestPtrSet = make(map[string]bool)
	g.data.SrcSubTypeMap = make(map[string]string)
	g.data.DestSubTypeMap = make(map[string]string) //qualified
	g.data.MismatchSubListMap = make(map[string]string)
	g.data.DestMismatchSubListMap = make(map[string]string)

	g.writeSrcMap = make(map[string]string)
	g.readSrcMap = make(map[string]string)

	for _, f1 := range g.exportedFields {
		for _, f2 := range g.destExportedFields {
			if !canNameMatch(f1, f2, g.tagMap) {
				continue
			}

			same, conv := matchType(f1.typ, f2.typ)
			if same || conv {
				continue
			}

			g.makeFuncMap(f1, f2)
			g.makeSubMap(f1, f2)
			g.makeSubListMap(f1, f2)
		}
	}
}

func (g *Generator) makeFuncMap(f1, f2 Field) {
	for _, fn := range g.mappingFuncList {
		if !g.writeDestSet[f2.name] && !f2.isGet {
			//in ToXxx, mapping func's param type is src field type
			if fn.param.String() == f1.typ.String() && fn.result.String() == f2.typ.String() {
				g.data.SrcToDestFuncMap[f1.name] = fn.name
				g.data.MismatchFuncMap[f1.name] = f2.name
				g.writeDestSet[f2.name] = true
				g.readSrcMap[f1.name] = f2.name
			}
		}

		if !g.writeSrcSet[f1.name] && !f1.isGet {
			//in FromXxx, mapping func's param type is dest field type
			if fn.param.String() == f2.typ.String() && fn.result.String() == f1.typ.String() {
				g.data.DestToSrcFuncMap[f1.name] = fn.name
				g.data.MismatchFuncMap[f1.name] = f2.name
				g.writeSrcSet[f1.name] = true
				g.writeSrcMap[f1.name] = f2.name
			}
		}
	}
}

func (g *Generator) makeSubMap(sub1, sub2 Field) {
	typ1 := sub1.typ
	typ2 := sub2.typ
	if p, ok := typ1.(*types.Pointer); ok {
		g.data.SrcPtrSet[sub1.name] = true
		typ1 = p.Elem()
	}

	if p, ok := typ2.(*types.Pointer); ok {
		g.data.DestPtrSet[sub2.name] = true
		typ2 = p.Elem()
	}

	if n1, ok := typ1.(*types.Named); ok {
		pkgpath1 := n1.Obj().Pkg().Path()
		g.data.SrcSubTypeMap[sub1.name] = n1.Obj().Name()

		if n2, ok := typ2.(*types.Named); ok {
			pkgpath2 := n2.Obj().Pkg().Path()
			g.data.DestSubTypeMap[sub2.name] = qualifiedTypeName(typ2, g.flags.alias)

			if pkgpath1 == g.Pkg().PkgPath && pkgpath2 == g.destPkg.PkgPath {
				if !g.writeSrcSet[sub1.name] && !sub1.isGet {
					g.data.MismatchSubMap[sub1.name] = sub2.name
					g.writeSrcSet[sub1.name] = true
					g.writeSrcMap[sub1.name] = sub2.name
				}
				if !g.writeDestSet[sub2.name] && !sub2.isGet {
					g.data.DestMismatchSubMap[sub1.name] = sub2.name
					g.writeDestSet[sub2.name] = true
					g.readSrcMap[sub1.name] = sub2.name
				}
			}
		}
	}
}

func (g *Generator) makeSubListMap(subs1, subs2 Field) {
	var typ1, typ2 types.Type
	if s, ok := subs1.typ.(*types.Slice); ok {
		typ1 = s.Elem()
		if p, ok := typ1.(*types.Pointer); ok {
			g.data.SrcPtrSet[subs1.name] = true
			typ1 = p.Elem()
		}
	} else {
		return
	}
	if s, ok := subs2.typ.(*types.Slice); ok {
		typ2 = s.Elem()
		if p, ok := typ2.(*types.Pointer); ok {
			g.data.DestPtrSet[subs2.name] = true
			typ2 = p.Elem()
		}
	} else {
		return
	}

	if n1, ok := typ1.(*types.Named); ok {
		pkgpath1 := n1.Obj().Pkg().Path()
		g.data.SrcSubTypeMap[subs1.name] = n1.Obj().Name()

		if named2, ok := typ2.(*types.Named); ok {
			pkgpath2 := named2.Obj().Pkg().Path()
			g.data.DestSubTypeMap[subs2.name] = qualifiedTypeName(typ2, g.flags.alias)

			if pkgpath1 == g.Pkg().PkgPath && pkgpath2 == g.destPkg.PkgPath {
				if !g.writeSrcSet[subs1.name] && !subs1.isGet {
					g.data.MismatchSubListMap[subs1.name] = subs2.name
					g.writeSrcSet[subs1.name] = true
					g.writeSrcMap[subs1.name] = subs2.name
				}
				if !g.writeDestSet[subs2.name] && !subs2.isGet {
					g.data.DestMismatchSubListMap[subs1.name] = subs2.name
					g.writeDestSet[subs2.name] = true
					g.readSrcMap[subs1.name] = subs2.name
				}
			}
		}
	}
}

func findImportForSelector(file *ast.File, sel *ast.SelectorExpr) *ast.ImportSpec {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil
	}

	name := ident.Name
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == name {
				return imp
			}
		} else {
			path := strings.Trim(imp.Path.Value, `"`)
			base := filepath.Base(path)
			if base == name {
				return imp
			}
		}
	}
	return nil
}
