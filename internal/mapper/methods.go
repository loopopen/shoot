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

func (g *Generator) parseGetSetIface(pkg *packages.Package, stType types.Type, typeName string, funcs *[]Func) {
	ptrType := types.NewPointer(stType)
	get, set := findGetterSetterIfac(pkg, typeName)
	if get == nil && set == nil {
		return
	}

	if named, ok := assignableToIface(ptrType, get); ok {
		iface, ok := named.Underlying().(*types.Interface)
		if ok {
			for i := 0; i < iface.NumMethods(); i++ {
				fn := iface.Method(i)
				sig := fn.Type().(*types.Signature)
				params := sig.Params()
				results := sig.Results()
				if params != nil && params.Len() > 0 {
					continue
				}
				if results == nil || results.Len() == 0 || results.Len() > 1 {
					continue
				}
				*funcs = append(*funcs, Func{
					name:   fn.Name(),
					result: results.At(0).Type(),
					path:   fn.Name(),
				})
			}
		}
	}
	if named, ok := assignableToIface(ptrType, set); ok {
		iface, ok := named.Underlying().(*types.Interface)
		if ok {
			for i := 0; i < iface.NumMethods(); i++ {
				fn := iface.Method(i)
				sig := fn.Type().(*types.Signature)
				params := sig.Params()
				results := sig.Results()
				if results != nil && results.Len() > 0 {
					continue
				}
				if params == nil || params.Len() == 0 || params.Len() > 1 {
					continue
				}
				*funcs = append(*funcs, Func{
					name:  fn.Name(),
					param: params.At(0).Type(),
					path:  fn.Name(),
				})
			}
		}
	}
}

func (g *Generator) parseGetSetMethods(pkg *packages.Package, stTyp types.Type, unexportedFields []*Field, funcs *[]Func) {
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
	}
}

func (g *Generator) parseSrcGetSetMethods(srcTyp types.Type, typeName string) {
	g.getsetMethods = nil
	g.parseGetSetIface(g.Pkg(), srcTyp, typeName, &g.getsetMethods)
	if len(g.getsetMethods) == 0 {
		g.parseGetSetMethods(g.Pkg(), srcTyp, g.unexportedFields, &g.getsetMethods)
	}
}

func (g *Generator) parseDestGetSetMethods(destTyp types.Type, typeName string) {
	g.destGetSetMethods = nil
	g.parseGetSetIface(g.destPkg, destTyp, typeName, &g.destGetSetMethods)
	if len(g.destGetSetMethods) == 0 {
		g.parseGetSetMethods(g.destPkg, destTyp, g.destUnexportedFields, &g.destGetSetMethods)
	}
}

func compatlize(fields *[]*Field, methods []Func) {
	for _, m := range methods {
		if m.param != nil { //set
			*fields = append(*fields, &Field{
				Name:        m.name,
				path:        m.path,
				typ:         m.param,
				backingName: trimLeftOnce(m.name, set),
				IsSet:       true,
			})
		} else if m.result != nil { //get
			*fields = append(*fields, &Field{
				Name:        m.name,
				path:        m.path,
				typ:         m.result,
				backingName: m.name,
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

func findGetterSetterIfac(pkg *packages.Package, name string) (types.Type, types.Type) {
	getterName := name + "Getter"
	setterName := name + "Setter"
	var get, set types.Type

	scope := pkg.Types.Scope()

	if obj := scope.Lookup(getterName); obj != nil {
		if typeName, ok := obj.(*types.TypeName); ok {
			get = typeName.Type()
		}
	}

	if obj := scope.Lookup(setterName); obj != nil {
		if typeName, ok := obj.(*types.TypeName); ok {
			set = typeName.Type()
		}
	}
	return get, set
}
