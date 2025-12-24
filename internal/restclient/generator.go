package restclient

import (
	_ "embed"
	"flag"
	"go/ast"
	"go/types"

	"github.com/lopolopen/shoot/internal/shoot"
)

const SubCmd = "rest"

//go:embed restclient.tmpl
var tmplTxt string

type Generator struct {
	*shoot.GeneratorBase
	// flags *Flags
	data *TmplData
}

func New() *Generator {
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
	}
	return g
}

func (g *Generator) ParseFlags() {
	sub := flag.NewFlagSet(SubCmd, flag.ExitOnError)

	g.ParseCommonFlags(sub)
}

func (g *Generator) MakeData(typeName string) any {
	g.data = NewTmplData(
		g.CommonFlags().CmdLine,
		g.CommonFlags().Version,
	)
	g.cookClient(typeName)
	g.data.SetTypeName(typeName)
	g.data.SetPackageName(g.Pkg().Name)
	return g.data
}

func (g *Generator) ListTypes() []string {
	var typeNames []string
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if !g.testNode("", n) {
				return true
			}

			ts, _ := n.(*ast.TypeSpec)
			typeNames = append(typeNames, ts.Name.Name)
			return false
		})
	}
	return typeNames
}

func (g *Generator) testNode(typeName string, node ast.Node) bool {
	ts, ok := node.(*ast.TypeSpec)
	if !ok {
		return false
	}

	if typeName != "" && ts.Name.Name != typeName {
		return false
	}

	iface, ok := ts.Type.(*ast.InterfaceType)
	if !ok {
		return false
	}

	for _, field := range iface.Methods.List {
		if len(field.Names) > 0 {
			continue
		}

		typ := g.Pkg().TypesInfo.Types[field.Type].Type
		named, ok := typ.(*types.Named)
		if !ok {
			continue
		}
		obj := named.Obj()
		pkgPath := obj.Pkg().Path()
		if pkgPath == shoot.SelfPkgPath && obj.Name() == "RestClient" {
			return true
		}
	}
	return false
}
