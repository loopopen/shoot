package mapper

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
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

	var expFuncList []shoot.Func
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

					expFuncList = append(expFuncList, shoot.Func{
						Name:   fn.Name.Name,
						Param:  g.mapperpkg.TypesInfo.TypeOf(params[0].Type),
						Result: g.mapperpkg.TypesInfo.TypeOf(results[0].Type),
					})
				}
			}

			return false
		})
	}
	g.mappingFuncList = expFuncList
}

func (g *Generator) makeTypeMismatch() {
	g.writeSrcMap = make(map[string]string)
	g.readSrcMap = make(map[string]string)

	for _, f1 := range g.exportedFields {
		for _, f2 := range g.destExportedFields {
			if !canNameMatch(f1, f2, g.srcTagMap, g.flags.ignoreCase) {
				continue
			}

			g.makeFuncMap(f1, f2)
			g.makeSubMap(f1, f2, f1.typ, f2.typ, false)
			g.makeSubListMap(f1, f2)
		}
	}
}

func (g *Generator) makeFuncMap(f1, f2 *Field) {
	for _, fn := range g.mappingFuncList {
		if !g.writeDestSet.Has(f2.Name) && !f2.IsGet {
			//in ToXxx, mapping func's param type is src field type
			if shoot.TypeEquals(fn.Param, f1.typ) && shoot.TypeEquals(fn.Result, f2.typ) {
				f1.Target = f2
				f2.Func = fn.Name
				g.writeDestSet.Adds(f2.Name)
				g.readSrcMap[f1.Name] = f2.Name
			}
		}

		if !g.writeSrcSet.Has(f1.Name) && !f1.IsGet {
			//in FromXxx, mapping func's param type is dest field type
			if shoot.TypeEquals(fn.Param, f2.typ) && shoot.TypeEquals(fn.Result, f1.typ) {
				f2.Target = f1
				f1.Func = fn.Name
				g.writeSrcSet.Adds(f1.Name)
				g.writeSrcMap[f1.Name] = f2.Name
			}
		}

		if f1.Target != nil && f2.Target != nil {
			break
		}
	}
}

func (g *Generator) makeSubMap(f1, f2 *Field, typ1, typ2 types.Type, isSlice bool) {
	var isPtr1, isPtr2 bool
	if p, ok := typ1.(*types.Pointer); ok {
		isPtr1 = true
		typ1 = p.Elem()
	}

	if p, ok := typ2.(*types.Pointer); ok {
		isPtr2 = true
		typ2 = p.Elem()
	}

	if n1, ok := typ1.(*types.Named); ok {
		pkgpath1 := n1.Obj().Pkg().Path()

		if n2, ok := typ2.(*types.Named); ok {
			pkgpath2 := n2.Obj().Pkg().Path()

			if pkgpath1 == g.Pkg().PkgPath && pkgpath2 == g.destPkg.PkgPath {
				if !g.writeDestSet.Has(f2.Name) && !f2.IsGet {
					//f2 = f1.ToDest()
					f1.Target = f2
					if isSlice {
						f2.CanEachMap = true
					} else {
						f2.CanMap = true
					}
					f2.Type = qualifiedTypeName(typ2, g.flags.alias)
					f1.IsPtr = isPtr1
					f2.IsPtr = isPtr2
					g.writeDestSet.Adds(f2.Name)
					g.readSrcMap[f1.Name] = f2.Name
				}
				if !g.writeSrcSet.Has(f1.Name) && !f1.IsGet {
					f2.Target = f1
					if isSlice {
						f1.CanEachMap = true
					} else {
						f1.CanMap = true
					}
					f1.Type = n1.Obj().Name()
					f1.IsPtr = isPtr1
					f2.IsPtr = isPtr2
					g.writeSrcSet.Adds(f1.Name)
					g.writeSrcMap[f1.Name] = f2.Name
				}
			}
		}
	}
}

func (g *Generator) makeSubListMap(f1, f2 *Field) {
	s1, ok := f1.typ.(*types.Slice)
	if !ok {
		return
	}
	s2, ok := f2.typ.(*types.Slice)
	if !ok {
		return
	}
	g.makeSubMap(f1, f2, s1.Elem(), s2.Elem(), true)
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
