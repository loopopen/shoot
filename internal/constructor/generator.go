package constructor

import (
	_ "embed"
	"flag"
	"go/ast"

	"github.com/lopolopen/shoot/internal/shoot"
)

const SubCmd = "new"

//go:embed constructor.tmpl
var tmplTxt string

// Generator holds the state of the analysis.
type Generator struct {
	*shoot.GeneratorBase
	flags *Flags
	data  *TmplData
}

func New() *Generator {
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
	}
	return g
}

func (g *Generator) ParseFlags() {
	sub := flag.NewFlagSet(SubCmd, flag.ExitOnError)
	getset := sub.Bool("getset", false, "generate Get/Set method for the type")
	json := sub.Bool("json", false, "generate MarshalJSON/UnmarshalJSON method for the type")
	var tagcase TagCase
	sub.Var(&tagcase, "tagcase", "specify the case style for struct tags (pascal, camel, lower, upper)")
	option := sub.Bool("option", false, "generate functional option pattern constructor")
	opt := sub.Bool("opt", false, "generate functional option pattern constructor (alias for -option)")
	exported := sub.Bool("exported", false, "“constructor parameters include exported fields")
	exp := sub.Bool("exp", false, "“constructor parameters include exported fields (alias for -exported)")
	short := sub.Bool("short", false, "shorter config function name (no 'OfType' suffix)")

	g.ParseCommonFlags(sub)

	g.flags = &Flags{
		getset:  *getset,
		json:    *json,
		tagcase: tagcase.String(),
		opt:     *opt || *option,
		exp:     *exp || *exported,
		short:   *short,
	}
}

func (g *Generator) MakeData(typeName string) any {
	g.data = NewTmplData(
		g.CommonFlags().CmdLine,
		g.CommonFlags().Version,
	)
	g.makeNew(typeName)
	g.makeOpt(typeName)
	g.makeGetSet(typeName)
	g.makeJson(typeName)
	g.data.SetTypeName(typeName)
	g.data.SetPackageName(g.Pkg().Name)
	return g.data
}

func (g *Generator) ListTypes() []string {
	var typeNames []string
	for _, f := range g.Pkg().Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			_, ok = ts.Type.(*ast.StructType)
			if !ok {
				return true
			}

			typeNames = append(typeNames, ts.Name.Name)
			return false
		})
	}
	return typeNames
}
