package mapper

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

func (g *Generator) parseCtors(srcTyp, destTyp types.Type, srcTypName, destTypName string) {
	shootnewIface := g.newShooterIface()
	if types.AssignableTo(srcTyp, shootnewIface) {
		fields := parseCtors(g.Pkg(), srcTyp, srcTypName)
		g.srcCtorParams = fields
	}

	if types.AssignableTo(destTyp, shootnewIface) {
		fields := parseCtors(g.destPkg, destTyp, destTypName)
		g.destCtorParams = fields
	}
}

func parseCtors(pkg *packages.Package, theTyp types.Type, typName string) []*Field {
	ctorName := "New" + typName
	var fields []*Field
	for _, f := range pkg.Syntax {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv != nil {
				continue
			}
			if fn.Name.Name != ctorName {
				continue
			}
			if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 || len(fn.Type.Results.List) > 1 {
				continue
			}
			r := fn.Type.Results.List[0]
			rExpr, ok := r.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			rTyp := pkg.TypesInfo.TypeOf(rExpr.X)
			if !types.Identical(rTyp, theTyp) {
				continue
			}

			params := fn.Type.Params
			if params == nil || len(params.List) == 0 {
				continue
			}

			nameMap := extractParamToFieldMap(fn)
			for _, p := range params.List {
				pname := p.Names[0].Name //params: camelCase

				n := nameMap[pname]
				name := n.name
				backing := name
				if !ast.IsExported(name) {
					name = set + transfer.ToPascalCase(name) //ref:02
				}

				fields = append(fields, &Field{
					Name:        name,
					backingName: backing,
					Path:        n.path,
					typ:         pkg.TypesInfo.TypeOf(p.Type),
				})
			}
		}
	}
	return fields
}

func (g *Generator) makeCtorMatch() {
	hasNonZero := makeCtorMatch(g, g.exportedFields, g.destCtorParams, g.srcTagMap, g.writeDestSet)
	if hasNonZero {
		g.data.DestCtorParams = g.destCtorParams
	}
	hasNonZero = makeCtorMatch(g, g.destExportedFields, g.srcCtorParams, nil, g.writeSrcSet)
	if hasNonZero {
		g.data.SrcCtorParams = g.srcCtorParams
	}
}

func makeCtorMatch(g *Generator, expFields []*Field, ctorParams []*Field, tagMap map[string]string, writeSet map[string]bool) bool {
	if len(ctorParams) == 0 {
		return false
	}
	hasNonZero := false

	for _, f := range expFields {
		for _, p := range ctorParams {
			if !canNameMatch(f, p, tagMap, g.flags.ignoreCase) {
				continue
			}

			if writeSet[p.Name] {
				continue
			}

			same, conv := matchType(f.typ, p.typ)
			//NewDest(f)
			//NewDest((type)f, ...)
			if same {
				p.CanAssign = true
			} else if conv {
				p.IsConv = true
				p.Type = qualifiedTypeName(p.typ, g.flags.alias)
			}
			if same || conv {
				p.Target = f //ref:01; Target has different meanings
				writeSet[p.Name] = true
				continue
			}

			for _, fn := range g.mappingFuncList {
				if shoot.TypeEquals(fn.Param, f.typ) && shoot.TypeEquals(fn.Result, p.typ) {
					p.Target = f
					p.Func = fn.Name
					writeSet[p.Name] = true
					continue
				}
			}

			//todo: ...
		}
	}
	for _, f2 := range ctorParams {
		if f2.Target != nil {
			hasNonZero = true
			continue
		}
		zero := zeroValue(f2.typ, g.qualifier)
		if zero == "" {
			//todo:
			logx.Fatal("not supported")
		}
		f2.Zero = zero
	}
	return hasNonZero
}

func zeroValue(t types.Type, qf func(*types.Package) string) string {
	switch tt := t.(type) {
	case *types.Basic:
		switch tt.Kind() {
		case types.String:
			return `""`
		case types.Bool:
			return "false"
		default:
			return "0"
		}

	case *types.Pointer:
		return "nil"
	case *types.Slice:
		return "nil"
	case *types.Map:
		return "nil"
	case *types.Chan:
		return "nil"
	case *types.Signature:
		return "nil"
	case *types.Interface:
		return "nil"

	case *types.Array:
		elemType := types.TypeString(tt.Elem(), qf)
		return fmt.Sprintf("[%d]%s{}", tt.Len(), elemType)

	case *types.Struct:
		if tt.NumFields() == 0 {
			return "struct{}{}"
		}
		return types.TypeString(tt, qf) + "{}"

	case *types.Named:
		under := tt.Underlying()
		if _, ok := under.(*types.Struct); ok {
			return types.TypeString(tt, qf) + "{}"
		}
		return zeroValue(under, qf)
	default:
		return ""
	}
}

type name struct {
	name string
	path string
}

func extractParamToFieldMap(fn *ast.FuncDecl) map[string]name {
	m := make(map[string]name)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) == 0 {
			return true
		}
		u, ok := ret.Results[0].(*ast.UnaryExpr)
		if !ok {
			return true
		}
		cl, ok := u.X.(*ast.CompositeLit)
		if !ok {
			return true
		}
		extractFromCompositeLit(cl, "", m)
		return true
	})
	return m
}

func extractFromCompositeLit(cl *ast.CompositeLit, prefix string, out map[string]name) {
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		path := key.Name
		if prefix != "" {
			path = prefix + "." + path
		}
		switch v := kv.Value.(type) {
		case *ast.Ident:
			out[v.Name] = name{
				name: key.Name,
				path: path,
			}
		case *ast.CompositeLit:
			extractFromCompositeLit(v, path, out)
		}
	}
}
