package mapper

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
)

func (g *Generator) parseManual(srcTypeName, destTypeName string) []string {
	g.writeSrcSet = make(map[string]bool)
	g.writeDestSet = make(map[string]bool)

	pkg := g.Pkg()
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode(srcTypeName, n) {
				return true
			}

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

					var recvTypeName string
					recv := fn.Recv.List[0]
					switch expr := recv.Type.(type) {
					case *ast.Ident: //value receiver
						if expr.Name != srcTypeName {
							continue
						}
						recvTypeName = srcTypeName
					case *ast.StarExpr: //pointer receiver
						if ident, ok := expr.X.(*ast.Ident); ok && ident.Name != srcTypeName {
							continue
						}
						recvTypeName = star + srcTypeName
					}

					if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
						continue
					}
					if fn.Type.Results != nil && len(fn.Type.Results.List) != 0 {
						continue
					}

					paramTypeExpr := fn.Type.Params.List[0].Type
					switch expr := paramTypeExpr.(type) {
					case *ast.Ident:
						//todo:
					case *ast.SelectorExpr: //value dest param
						paramTypeExpr = expr.Sel
						if isWrite {
							logx.Fatalf("method (%s).%s must has a pointer parameter", recvTypeName, fn.Name.Name)
						} else {
							g.data.ReadParamPrefix = star
						}
					case *ast.StarExpr: //pointer dest param
						paramTypeExpr = expr.X
					}

					//checking param type needs full type name
					destFullName := fmt.Sprintf("%s.%s", g.destPkg.PkgPath, destTypeName)
					paramType := pkg.TypesInfo.TypeOf(paramTypeExpr)
					if paramType.String() != destFullName {
						continue
					}

					if isWrite {
						if g.data.WriteMethodName == "" {
							g.data.WriteMethodName = fn.Name.Name
						} else {
							logx.Fatalf("found more than one manual write method: (%s).%s", recvTypeName, fn.Name.Name)
						}
						names := findAssignedFieldPaths(fn, fn.Type.Params.List[0].Names[0].Name)
						for _, n := range names {
							g.writeDestSet[n] = true
						}
					} else if isRead {
						if g.data.ReadMethodName == "" {
							g.data.ReadMethodName = fn.Name.Name
						} else {
							logx.Fatalf("found more than one manual read method: (%s).%s", recvTypeName, fn.Name.Name)
						}
						names := findAssignedFieldPaths(fn, recv.Names[0].Name)
						for _, n := range names {
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
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			path := findPath(lhs, v)
			if path != "" {
				fieldPath = append(fieldPath, path)
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
