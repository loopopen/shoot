package enumer

import (
	_ "embed"
	"flag"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"golang.org/x/tools/go/packages"
)

const SubCmd = "enum"

//go:embed enumer.tmpl
var tmplTxt string

// Generator holds the state of the analysis.
type Generator struct {
	*shoot.GeneratorBase
	flags *Flags
	data  *TmplData
	pkg   *Package
}

func New() *Generator {
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt)}
	return g
}

func (g *Generator) ParseFlags() {
	sub := flag.NewFlagSet(SubCmd, flag.ExitOnError)
	bit := sub.Bool("bit", false, "generate bitwise enumerations (alias for -bitwise)")
	bitwise := sub.Bool("bitwise", false, "generate bitwise enumerations")
	json := sub.Bool("json", false, "generate MarshalJSON/UnmarshalJSON method for the type")
	text := sub.Bool("text", false, "generate MarshaText/UnmarshalText method for the type")
	sql := sub.Bool("sql", false, "generate Value/Scan method for the type")
	gorm := sub.Bool("gorm", false, "generate GormDataType/GormDBDataType for the type")

	g.ParseCommonFlags(sub)

	if *gorm && !*sql {
		logx.Fatal("-gorm only works when -sql is enabled")
	}

	g.flags = &Flags{
		bitwise: *bitwise || *bit,
		json:    *json,
		text:    *text,
		sql:     *sql,
		gorm:    *gorm,
	}
}

func (g *Generator) LoadPackage(patterns ...string) map[string]*packages.Package {
	pkgs := g.GeneratorBase.LoadPackage(patterns...)
	g.addPackage(g.Pkg())
	return pkgs
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		pkg:   pkg,
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

func (g *Generator) MakeData(typeName string) (any, bool) {
	g.data = NewTmplData(
		g.CommonFlags().CmdLine,
		g.CommonFlags().Version,
	)
	g.makeStr(typeName)
	g.makeBitwize()
	g.makeJson()
	g.makeText()
	g.makeSQL()

	if len(g.data.NameList) == 0 {
		return nil, false
	}

	g.data.SetTypeName(typeName)
	g.data.SetPackageName(g.Pkg().Name)
	return g.data, false
}

func (g *Generator) ListTypes() []string {
	var typeNames []string
	pkg := g.Pkg()
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok {
				return true
			}

			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					obj := pkg.TypesInfo.Defs[ts.Name]
					if obj == nil {
						continue
					}

					typ := obj.Type()
					under := typ.Underlying()

					basic, ok := under.(*types.Basic)
					if !ok {
						continue
					}

					kind := basic.Kind()
					if kind != types.Int && kind != types.Uint &&
						kind != types.Int32 && kind != types.Uint32 {
						continue
					}

					if ts.Assign.IsValid() {
						logx.Warnf("alias type %s will be ignored", ts.Name.Name)
					} else {
						typeNames = append(typeNames, ts.Name.Name)
					}
				}
			}

			return false
		})
	}
	return typeNames
}
