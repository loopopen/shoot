package mapper

import (
	_ "embed"
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/lopolopen/shoot/internal/shoot"
	"github.com/lopolopen/shoot/internal/tools/logx"
	"golang.org/x/tools/go/packages"
)

const SubCmd = "map"

const (
	dot  = "."
	star = "*"
	set  = "Set"
)

//go:embed mapper.tmpl
var tmplTxt string

// Generator holds the state of the analysis.
type Generator struct {
	*shoot.GeneratorBase
	flags                *Flags
	data                 *TmplData
	destPkg              *packages.Package
	mapperpkg            *packages.Package
	exportedFields       []*Field
	unexportedFields     []*Field
	destExportedFields   []*Field
	destUnexportedFields []*Field

	getsetMethods     []Func
	destGetSetMethods []Func

	srcPtrTypeMap   map[string]string
	destPtrTypeMap  map[string]string
	srcPathsMap     map[string][]string
	destPathsMap    map[string][]string
	mappingFuncList []Func

	writeSrcSet  map[string]bool
	writeDestSet map[string]bool
	readSrcMap   map[string]string
	writeSrcMap  map[string]string

	srcTagMap map[string]string
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
	if pkg.Path() == g.destPkg.PkgPath {
		if g.flags.alias != "" {
			return g.flags.alias
		}
	}
	return pkg.Name()
}

func (g *Generator) ParseFlags() {
	sub := flag.NewFlagSet(SubCmd, flag.ExitOnError)
	path := sub.String("path", "", "destination package path to map to")
	alias := sub.String("alias", "", "destination package alias")
	destTypes := sub.String("to", "", "destination type names to map to (must align to -type)")
	var way Way
	sub.Var(&way, "way", "limit the mapping way(toonly/->, fromonly/<-, both/<->)")

	g.ParseCommonFlags(sub)

	if way == "" {
		way = WayBoth
	}

	typNames := g.CommonFlags().TypeNames
	typMap := make(map[string]string)
	if len(typNames) == 0 {
		if *destTypes != "" {
			logx.Fatal("-to only works when -type used")
		}
	} else {
		destTypNames := strings.Split(*destTypes, ",")
		if *destTypes != "" && len(destTypNames) != len(typNames) {
			logx.Fatal("-to list must align to -type list")
		}
		if *destTypes == "" {
			for i, n := range typNames {
				typMap[n] = typNames[i]
			}
		} else {
			for i, n := range typNames {
				typMap[n] = destTypNames[i]
			}
		}
	}

	destDir := shoot.FixPath(*path)
	_, err := os.Stat(filepath.Join(g.CommonFlags().Dir, destDir))
	if err != nil && !os.IsExist(err) {
		logx.Fatalf("destination dir not exists: %s", destDir)
	}

	g.flags = &Flags{
		destDir:   destDir,
		destTypes: typMap,
		alias:     *alias,
		way:       way,
	}
}

func (g *Generator) LoadPackage(patterns ...string) map[string]*packages.Package {
	dest := g.flags.destDir
	patterns = append(patterns, dest)
	pkgs := g.GeneratorBase.LoadPackage(patterns...)
	g.destPkg = pkgs[dest]
	return pkgs
}

func (g *Generator) loadMorePkgs(srcTypeName string) {
	g.mappingFuncList = nil
	mapperTypeName := g.loadTypeMapperPkg(srcTypeName)
	if mapperTypeName != "" {
		g.parseMapper(mapperTypeName)
	}
}

func (g *Generator) MakeData(srcTypeName string) any {
	//will reload and reset all pkgs, call as early as possible
	g.loadMorePkgs(srcTypeName)

	g.data = NewTmplData(
		g.CommonFlags().CmdLine,
		g.CommonFlags().Version,
	)

	var destTypeName string
	if g.flags.destTypes != nil {
		if name, ok := g.flags.destTypes[srcTypeName]; ok {
			destTypeName = name
		}
	}
	if destTypeName == "" {
		destTypeName = srcTypeName
	}
	g.data.DestTypeName = destTypeName

	shootnewIfac := makeNewShooterIfac()
	srcTyp := g.parseSrcFields(srcTypeName)
	if types.AssignableTo(srcTyp, shootnewIfac) {
		g.parseSrcGetSetMethods(srcTyp, srcTypeName)
	}
	if srcTyp == nil {
		logx.Fatalf("src type not exists: %s", srcTypeName)
	}
	destTyp := g.parseDestFields(destTypeName)
	if destTyp == nil {
		if g.IsTypeSpecified() {
			logx.Fatalf("dest type not exists: %s", destTypeName)
		} else {
			return nil
		}
	}
	if types.AssignableTo(destTyp, shootnewIfac) {
		g.parseDestGetSetMethods(destTyp, destTypeName)
	}

	g.parseManual(srcTyp, destTyp)
	g.makeCompatible()
	g.makeMismatch() //priority: makeMismatch > makeMatch
	g.makeMatch()

	g.data.DestPkgName = g.destPkg.Name
	g.data.DestPkgPath = g.destPkg.PkgPath
	g.data.DestPkgAlias = g.flags.alias
	g.data.QualifiedDestTypeName = types.TypeString(destTyp, g.qualifier)
	g.data.IsToOnly = g.flags.way == WayToOnly
	g.data.IsFromOnly = g.flags.way == WayFromOnly
	g.data.SetTypeName(srcTypeName)
	g.data.SetPackageName(g.Pkg().Name)

	g.makeReadWriteCheck()
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

func (g *Generator) testNode(srcType string, node ast.Node) bool {
	ts, ok := node.(*ast.TypeSpec)
	if !ok {
		return false
	}

	if srcType != "" && ts.Name.Name != srcType {
		return false
	}

	_, ok = ts.Type.(*ast.StructType) //empty struct is ok
	if !ok {
		return false
	}

	if srcType == "" && !ast.IsExported(ts.Name.Name) {
		return false
	}

	return true
}

func (g *Generator) readDestMap() map[string]string {
	return g.writeSrcMap
}

func (g *Generator) writeDestMap() map[string]string {
	return g.readSrcMap
}

func makeNewShooterIfac() types.Type {
	iface := types.NewInterfaceType(
		[]*types.Func{
			types.NewFunc(
				token.NoPos,
				nil,
				"ShootNew",
				types.NewSignatureType(nil, nil, nil, nil, nil, false),
			),
		},
		nil,
	).Complete()
	return iface
}
