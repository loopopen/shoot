package mapper

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) parseMethods(srcTyp, destTyp types.Type, srcTypName, destTypName string) {
	shootnewIface := g.newShooterIface()
	if types.AssignableTo(srcTyp, shootnewIface) {
		g.parseSrcGetSetMethods(srcTyp, srcTypName)
	}

	if types.AssignableTo(destTyp, shootnewIface) {
		g.parseDestGetSetMethods(destTyp, destTypName)
	}
}

func (g *Generator) parseSrcGetSetMethods(srcTyp types.Type, typeName string) {
	g.getsetMethods = nil
	ptrTyp := types.NewPointer(srcTyp)
	shoot.ParseGetSetIface(g.Pkg(), ptrTyp, typeName, &g.getsetMethods)
	if len(g.getsetMethods) == 0 { //todo: need it?
		parseGetSetMethods(g.Pkg(), srcTyp, g.unexportedFields, &g.getsetMethods)
	}
}

func (g *Generator) parseDestGetSetMethods(destTyp types.Type, typeName string) {
	g.destGetSetMethods = nil
	ptrTyp := types.NewPointer(destTyp)
	shoot.ParseGetSetIface(g.destPkg, ptrTyp, typeName, &g.destGetSetMethods)
	if len(g.destGetSetMethods) == 0 {
		parseGetSetMethods(g.destPkg, destTyp, g.destUnexportedFields, &g.destGetSetMethods)
	}
}

func parseGetSetMethods(pkg *packages.Package, stTyp types.Type, unexportedFields []*Field, funcs *[]shoot.Func) {
	if len(unexportedFields) == 0 {
		return
	}

	super := make(map[string]*Field)
	for _, field := range unexportedFields {
		super[transfer.ToPascalCase(field.Name)] = field
		super[set+transfer.ToPascalCase(field.Name)] = field
	}

	for _, f := range pkg.Syntax {
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
				if !shoot.TypeEquals(paramTyp, field.typ) {
					continue
				}
				*funcs = append(*funcs, shoot.Func{
					Name:  fn.Name.Name,
					Param: paramTyp,
					Path:  fn.Name.Name,
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
				*funcs = append(*funcs, shoot.Func{
					Name:   fn.Name.Name,
					Result: resultTyp,
					Path:   fn.Name.Name,
				})
			}
		}
	}
}

func compatlize(fields *[]*Field, methods []shoot.Func) {
	for _, m := range methods {
		if m.Param != nil { //set
			*fields = append(*fields, &Field{
				Name:        m.Name,
				Path:        m.Path,
				typ:         m.Param,
				backingName: trimLeftOnce(m.Name, set),
				IsSet:       true,
			})
		} else if m.Result != nil { //get
			*fields = append(*fields, &Field{
				Name:        m.Name,
				Path:        m.Path,
				typ:         m.Result,
				backingName: m.Name,
				IsGet:       true,
			})
		}
	}
}

func (g *Generator) makeCompatible() {
	g.RegisterTransfer("dot", transfer.ID)
	g.RegisterTransfer("assign", transfer.ID)
	g.RegisterTransfer("evaluate", transfer.ID)

	compatlize(&g.exportedFields, g.getsetMethods)
	compatlize(&g.destExportedFields, g.destGetSetMethods)
	g.data.SrcFieldList = g.exportedFields
	g.data.DestFieldList = g.destExportedFields

	g.RegisterTransfer("assign", func(left, right string, isSet bool) string {
		//left ~ set
		//right ~ get

		assertArgs(left, "left")
		assertArgs(right, "right")

		if isSet {
			return fmt.Sprintf("%s(%s)", left, right)
		} else {
			return fmt.Sprintf("%s = %s", left, right)
		}
	})

	g.RegisterTransfer("dot", func(sel, x string) string {
		return fmt.Sprintf("%s.%s", sel, x)
	})

	//evaluation always happens on the right side
	g.RegisterTransfer("evaluate", func(rSel, rX string, isGet bool) string {
		if isGet {
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
