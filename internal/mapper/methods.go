package mapper

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) parseGetSetMethods(pkg *packages.Package, stTyp types.Type, unexportedFields []Field, funcs *[]Func) {
	if len(unexportedFields) == 0 {
		return
	}

	super := make(map[string]Field)
	for _, field := range unexportedFields {
		super[transfer.ToPascalCase(field.name)] = field
		super[set+transfer.ToPascalCase(field.name)] = field
	}

	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if fn.Recv == nil || len(fn.Recv.List) == 0 {
					continue
				}
				field, ok := super[fn.Name.Name]
				if !ok {
					continue
				}

				expr := fn.Recv.List[0].Type
				if ptr, ok := expr.(*ast.StarExpr); ok {
					expr = ptr.X
				}
				recvTyp := pkg.TypesInfo.TypeOf(expr)
				if !types.Identical(recvTyp, stTyp) {
					continue
				}

				params := fn.Type.Params
				results := fn.Type.Results
				if strings.HasPrefix(fn.Name.Name, set) {
					//set
					if results != nil && len(results.List) > 0 {
						continue
					}
					if params == nil || len(params.List) > 1 {
						continue
					}
					paramTyp := pkg.TypesInfo.TypeOf(params.List[0].Type)
					if !types.Identical(paramTyp, field.typ) {
						continue
					}
					*funcs = append(*funcs, Func{
						name:  fn.Name.Name,
						param: paramTyp,
						path:  fn.Name.Name,
					})
				} else {
					//get
					if params != nil && len(params.List) > 0 {
						continue
					}
					if results == nil || len(results.List) > 1 {
						continue
					}
					resultTyp := pkg.TypesInfo.TypeOf(results.List[0].Type)
					if !types.Identical(resultTyp, field.typ) {
						continue
					}
					*funcs = append(*funcs, Func{
						name:   fn.Name.Name,
						result: resultTyp,
						path:   fn.Name.Name,
					})
				}
			}

			return false
		})
	}
}

func (g *Generator) parseSrcGetSetMethods(srcTyp types.Type) {
	var methods []Func
	g.parseGetSetMethods(g.Pkg(), srcTyp, g.unexportedFields, &methods)
	for _, m := range methods {
		g.getsetMethods = append(g.getsetMethods, m)
	}
}

func (g *Generator) parseDestGetSetMethods(destTyp types.Type) {
	var methods []Func
	g.parseGetSetMethods(g.destPkg, destTyp, g.destUnexportedFields, &methods)
	for _, m := range methods {
		g.destGetSetMethods = append(g.destGetSetMethods, m)
	}
}

func (g *Generator) makeCompatible() {
	g.RegisterTransfer("assign", transfer.ID)
	g.RegisterTransfer("evaluate", transfer.ID)
	g.getSrcSet = make(map[string]bool)
	g.setSrcSet = make(map[string]bool)
	g.getDestSet = make(map[string]bool)
	g.setDestSet = make(map[string]bool)

	for _, m := range g.getsetMethods {
		if m.param != nil { //set
			g.exportedFields = append(g.exportedFields, Field{
				name:        m.name,
				path:        m.path,
				typ:         m.param,
				backingName: trimLeftOnce(m.name, set),
				isSet:       true,
			})
			g.setSrcSet[m.name] = true
		} else if m.result != nil { //get
			g.exportedFields = append(g.exportedFields, Field{
				name:        m.name,
				path:        m.path,
				typ:         m.result,
				backingName: m.name,
				isGet:       true,
			})
			g.getSrcSet[m.name] = true
		}
		g.data.SrcFieldList = append(g.data.SrcFieldList, m.name)
	}

	for _, m := range g.destGetSetMethods {
		if m.param != nil { //set
			g.destExportedFields = append(g.destExportedFields, Field{
				name:        m.name,
				path:        m.path,
				typ:         m.param,
				backingName: trimLeftOnce(m.name, set),
				isSet:       true,
			})
			g.setDestSet[m.name] = true
		} else if m.result != nil { //get
			g.destExportedFields = append(g.destExportedFields, Field{
				name:        m.name,
				path:        m.path,
				typ:         m.result,
				backingName: m.name,
				isGet:       true,
			})
			g.getDestSet[m.name] = true
		}
	}

	if g.CommonFlags().Verbose {
		logx.DebugJSON("Get of Src:\n", g.getSrcSet)
		logx.DebugJSON("Set of Src:\n", g.setSrcSet)
		logx.DebugJSON("Get of Dest:\n", g.getDestSet)
		logx.DebugJSON("Set of Dest:\n", g.setDestSet)
	}
	g.RegisterTransfer("assign", func(lSel, lX, right string, isLeftSrc bool, lfmt, rfmt string) string {
		//left ~ set
		//right ~ get

		// if g.CommonFlags().Verbose {
		// 	logx.Debug(fmt.Sprintf("%s.%s", l, lname))
		// 	logx.Debug(fmt.Sprintf("%s.%s", r, rname))
		// }

		assertArgs(lSel, "lSel")
		assertArgs(lX, "lX")
		assertArgs(right, "right")
		assertArgs(lfmt, "lfmt")
		assertArgs(rfmt, "rfmt")

		var setSet map[string]bool
		if isLeftSrc {
			setSet = g.setSrcSet
		} else {
			setSet = g.setDestSet
		}

		left := fmt.Sprintf(lfmt, lSel, lX)
		right = fmt.Sprintf(rfmt, right)
		if setSet[lX] {
			return fmt.Sprintf("%s(%s)", left, right)
		} else {
			return fmt.Sprintf("%s = %s", left, right)
		}
	})

	//evaluation always happens on the right side
	g.RegisterTransfer("evaluate", func(rSel, rX string, isSrc bool) string {
		var getSet map[string]bool
		if isSrc {
			getSet = g.getSrcSet
		} else {
			getSet = g.getDestSet
		}
		if getSet[rX] {
			rX = rX + "()"
		}
		return fmt.Sprintf("%s.%s", rSel, rX)
	})
}

func trimLeftOnce(s, cut string) string {
	if strings.HasPrefix(s, cut) {
		return s[len(cut):]
	}
	return s
}

func assertArgs(v, name string) {
	if v == "" {
		logx.Fatalf("%s should not be empty", name)
	}
}
