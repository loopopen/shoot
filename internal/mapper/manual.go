package mapper

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) parseManual(srcType, destType types.Type) []string {
	g.writeSrcSet = make(map[string]bool)
	g.writeDestSet = make(map[string]bool)

	pkg := g.Pkg()
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil {
					if len(fn.Recv.List) == 0 {
						continue
					}

					isWrite := isWriteMethod(fn.Name.Name, withAlias(g.destPkg.Name, g.flags.alias))
					isRead := !isWrite && isReadMethod(fn.Name.Name, withAlias(g.destPkg.Name, g.flags.alias))

					if !isWrite && !isRead {
						continue
					}

					const fmtIncorrectSig = "(%s).%s: signature of reserved func is incorrect, rename it or fix the signature"
					recv := fn.Recv.List[0]
					recvTypeExpr := recv.Type
					isRecvPtr := false
					switch expr := recv.Type.(type) {
					// case *ast.Ident: //value receiver
					case *ast.StarExpr: //pointer receiver
						recvTypeExpr = expr.X
						isRecvPtr = true
					default:
						panic("can never happen")
					}

					recvType := pkg.TypesInfo.TypeOf(recvTypeExpr)
					var recvTypeName string = types.TypeString(recvType, g.qualifier)
					if isRecvPtr {
						recvTypeName = star + recvTypeName
					}
					if !types.Identical(recvType, srcType) {
						continue
					}
					if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
						logx.Warnf(fmtIncorrectSig, recvTypeName, fn.Name.Name)
						continue
					}
					if fn.Type.Results != nil && len(fn.Type.Results.List) != 0 {
						logx.Warnf(fmtIncorrectSig, recvTypeName, fn.Name.Name)
						continue
					}

					param := fn.Type.Params.List[0]
					paramTypeExpr := param.Type
					isParamPtr := false
					switch expr := paramTypeExpr.(type) {
					// case *ast.Ident:
					case *ast.SelectorExpr: //value dest param
						paramTypeExpr = expr.Sel
					case *ast.StarExpr: //pointer dest param
						paramTypeExpr = expr.X
						isParamPtr = true
					}

					paramType := pkg.TypesInfo.TypeOf(paramTypeExpr)
					var isSame, isGetter, isSetter bool
					if types.Identical(paramType, destType) {
						isSame = true
					} else {
						if n, ok := paramType.(*types.Named); ok {
							if ifac, ok := n.Underlying().(*types.Interface); ok {
								if types.ConvertibleTo(types.NewPointer(destType), ifac) {
									name := n.Obj().Name()
									isGetter = strings.HasSuffix(name, "Getter")
									isSetter = strings.HasSuffix(name, "Setter")
								}
							}
						}
					}

					destTypeName := types.TypeString(destType, g.qualifier)
					if isWrite {
						if !(isSame && isParamPtr) && !isSetter {
							logx.Fatalf("(%s).%s: parameter of write method must be a pointer or a setter of %s", recvTypeName, fn.Name.Name, destTypeName)
						}

						if g.data.WriteMethodName == "" {
							g.data.WriteMethodName = fn.Name.Name
						} else {
							logx.Fatalf("found more than one manual write method: (%s).%s", recvTypeName, fn.Name.Name)
						}
						names := findAssignedFieldPaths(fn, param.Names[0].Name)
						for _, n := range names {
							g.writeDestSet[n] = true
						}
					} else if isRead { //write src
						if !isSame && !isGetter {
							logx.Fatalf("(%s).%s: parameter of read method must be a value or a getter of %s", recvTypeName, fn.Name.Name, destTypeName)
						}
						g.data.IsReadParamPtr = isParamPtr || isGetter

						if g.data.ReadMethodName == "" {
							g.data.ReadMethodName = fn.Name.Name
						} else {
							logx.Fatalf("found more than one manual read method: (%s).%s", recvTypeName, fn.Name.Name)
						}
						names := findAssignedFieldPaths(fn, recv.Names[0].Name)
						for _, n := range names {
							if types.ConvertibleTo(srcType, g.newShooterIface()) && !ast.IsExported(n) {
								//r.x = 0 => SetX, SetX may not exist
								n = set + transfer.ToPascalCase(n) //ref:02
							}
							g.writeSrcSet[n] = true
						}
					}
				}
			}
			return false
		})
	}
	return nil
}

func isWriteMethod(methodName string, destPkgKey string) bool {
	keys := []string{"to", "write"}
	for _, key := range keys {
		if methodName == key+strings.ToLower(destPkgKey) {
			return true
		}

		if methodName == key+transfer.ToPascalCase(destPkgKey) {
			return true
		}
	}
	return false
}

func isReadMethod(methodName string, destPkgName string) bool {
	keys := []string{"from", "read"}
	for _, key := range keys {
		if methodName == key+strings.ToLower(destPkgName) {
			return true
		}

		if methodName == key+transfer.ToPascalCase(destPkgName) {
			return true
		}
	}
	return false
}

func findAssignedFieldPaths(funcDecl *ast.FuncDecl, v string) []string {
	var fieldPath []string
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				path := findPath(lhs, v)
				if path != "" {
					fieldPath = append(fieldPath, path)
				}
			}
		} else if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if strings.HasPrefix(sel.Sel.Name, set) {
					path := findPath(sel, v)
					fieldPath = append(fieldPath, path)
				}
			}
		}
		return true
	})
	return fieldPath
}

func findPath(x ast.Expr, v string) string {
	var path string
	for {
		sel, ok := x.(*ast.SelectorExpr)
		if !ok {
			break
		}
		path = sel.Sel.Name + "." + path
		x = sel.X
	}
	if path == "" {
		return ""
	}
	i, ok := x.(*ast.Ident)
	if !ok || i.Name != v {
		return ""
	}
	return path[:len(path)-1]
}

func withAlias(name, alias string) string {
	if alias != "" {
		return alias
	}
	return name
}
