package constructor

import (
	_ "embed"
	"flag"
	"go/ast"
	"go/types"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
)

const SubCmd = "new"

//go:embed constructor.tmpl
var tmplTxt string

// Generator holds the state of the analysis.
type Generator struct {
	*shoot.GeneratorBase
	flags         *Flags
	data          *TmplData
	typeParams    []string
	typeParamsMap map[int]string
	fields        []*Field
	hasNew        bool
	getsetMethods []shoot.Func
	getter        bool
	setter        bool
}

func New() *Generator {
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
	}
	return g
}

func (g *Generator) qualifier(pkg *types.Package) string {
	if pkg == nil {
		return ""
	}
	if pkg.Path() == g.Pkg().PkgPath {
		return ""
	}
	return pkg.Name()
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

	if tagcase == "" {
		tagcase = TagCaseCamel
	}

	g.flags = &Flags{
		getset:  *getset,
		json:    *json,
		tagcase: tagcase,
		opt:     *opt || *option,
		exp:     *exp || *exported,
		short:   *short,
	}
}

func (g *Generator) MakeData(typeName string) (any, bool) {
	g.getter = true
	g.setter = true
	g.data = NewTmplData(
		g.CommonFlags().CmdLine,
		g.CommonFlags().Version,
	)

	theTyp := g.parseFields(typeName)
	if theTyp == nil {
		logx.Fatalf("type not exists: %s", typeName)
	}
	g.makeGetSet() //must be before makeNew & makeJson
	g.makeNew()
	g.makeJson()
	g.data.SetTypeName(typeName)
	g.data.SetPackageName(g.Pkg().Name)
	return g.data, g.data.GetSet
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

func (g *Generator) testNode(typename string, node ast.Node) bool {
	ts, ok := node.(*ast.TypeSpec)
	if !ok {
		return false
	}

	if typename != "" {
		if ts.Name.Name != typename {
			return false
		}

		if _, ok = ts.Type.(*ast.StructType); !ok {
			logx.Fatalf("type %s is not a struct type", typename)
		}
	}

	if strings.HasPrefix(ts.Name.Name, "_") {
		return false
	}

	if _, ok = ts.Type.(*ast.StructType); !ok {
		return false
	}

	return true
}
