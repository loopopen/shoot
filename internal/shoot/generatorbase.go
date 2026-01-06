package shoot

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/lopolopen/shoot/internal/tools/logx"
	"github.com/lopolopen/shoot/internal/transfer"
	"golang.org/x/tools/go/packages"
)

type GeneratorBase struct {
	commonFlags     *CommonFlags
	subCmd          string
	tmplTxt         string
	tmp             *template.Template
	transfers       template.FuncMap
	pkg             *packages.Package
	allInOneFile    string
	fileNameMap     map[string]string
	isTypeSpecified bool
}

func NewGeneratorBase(subCmd string, tmplTxt string) *GeneratorBase {
	g := &GeneratorBase{
		subCmd:      subCmd,
		tmplTxt:     tmplTxt,
		fileNameMap: make(map[string]string),
	}
	g.preRegister()
	return g
}

func (g *GeneratorBase) tmpl() *template.Template {
	if g.tmp == nil {
		tmp, err := template.New(g.subCmd).Funcs(g.transfers).Parse(g.tmplTxt)
		if err != nil {
			logx.Fatalf("parsing template: %s", err)
		}
		g.tmp = tmp
	}
	return g.tmp
}

func (g *GeneratorBase) CommonFlags() *CommonFlags {
	return g.commonFlags
}

func (g *GeneratorBase) Pkg() *packages.Package {
	return g.pkg
}

func (g *GeneratorBase) SetPkg(pkg *packages.Package) {
	g.pkg = pkg
}

func (g *GeneratorBase) IsTypeSpecified() bool {
	return g.isTypeSpecified
}

func (d *GeneratorBase) preRegister() {
	d.RegisterTransfer("firstLower", transfer.FirstLower)
	d.RegisterTransfer("camelCase", transfer.ToCamelCase)
	d.RegisterTransfer("pascalCase", transfer.ToPascalCase)
	d.RegisterTransfer("in", func(s string, list []string) bool {
		for _, x := range list {
			if s == x {
				return true
			}
		}
		return false
	})
	d.RegisterTransfer("cond", func(cond bool, a, b string) string {
		if cond {
			return a
		}
		return b
	})
}

func (d *GeneratorBase) RegisterTransfer(key string, transfer any) {
	if d.transfers == nil {
		d.transfers = make(template.FuncMap)
	}

	d.transfers[key] = transfer
	d.tmp = nil
}

func (g *GeneratorBase) ParseCommonFlags(sub *flag.FlagSet) {
	typeNames := sub.String("type", "", "comma-separated list of type names")
	filename := sub.String("file", "", "the targe go file to generate, typical value: $GOFILE")
	separate := sub.Bool("separate", false, "each type has its own go file")
	sep := sub.Bool("sep", false, "each type has its own go file (alias for separate)")
	verbose := sub.Bool("verbose", false, "verbose output")
	v := sub.Bool("v", false, "verbose output (alias for verbose)")
	raw := sub.Bool("raw", false, "raw source")
	r := sub.Bool("r", false, "raw source (alias for raw)")
	version := sub.String("version", "", "pin version")
	ver := sub.String("ver", "", "pin version (alias for version)")

	args := flag.Args()
	if len(args) <= 1 {
		sub.Usage()
		os.Exit(2)
	}

	cmdline := Shoot + " " + strings.Join(args, " ") //e.g.: shoot enum -bit type=YourType ./testdata
	sub.Parse(args[1:])
	if *typeNames == "" && *filename == "" {
		sub.Usage()
		os.Exit(2)
	}

	var types []string
	if *typeNames != "" {
		types = strings.Split(*typeNames, ",")
	}

	g.isTypeSpecified = *typeNames != "" && *typeNames != star
	sep_ := g.isTypeSpecified || *sep || *separate

	dir := FixPath(sub.Arg(0)) //e.g. ./testdata
	if dir != dot {
		_, err := os.Stat(dir)
		if err != nil && !os.IsExist(err) {
			logx.Fatalf("working dir not exists: %s", dir)
		}
	}

	if *filename != "" {
		if filepath.Ext(*filename) != ".go" {
			logx.Fatalf("file must be a go file: %s", *filename)
		}
		fp := filepath.Join(dir, *filename)
		_, err := os.Stat(fp)
		if err != nil && !os.IsExist(err) {
			logx.Fatalf("file not exists: %s", fp)
		}
	}

	if *version != "" {
		*ver = *version
	} else if *ver == "" {
		*ver = Version
	}

	g.commonFlags = &CommonFlags{
		CmdLine:   cmdline,
		TypeNames: types,
		FileName:  *filename,
		Separate:  sep_,
		Dir:       dir,
		Verbose:   *v || *verbose,
		Raw:       *r || *raw,
		Version:   *ver,
	}
}

func (g *GeneratorBase) fileName(typeName string, pkgScope bool) string {
	cmd := Shoot + g.subCmd
	if pkgScope {
		return fmt.Sprintf("%s.%s.go", cmd, typeName)
	}
	fileName := g.commonFlags.FileName
	if fileName == "" {
		fileName = g.allInOneFile
	}
	if fileName == "" {
		fileName = g.fileNameMap[typeName]
	}

	gofile := strings.TrimSuffix(fileName, ".go")
	if typeName == "" {
		return fmt.Sprintf("%s.%s.go", gofile, cmd)
	}
	if !ast.IsExported(typeName) {
		typeName = "_" + typeName
	}
	return fmt.Sprintf("%s.%s.%s.go", gofile, cmd, strings.ToLower(typeName))
}

func (g *GeneratorBase) LoadPackage(patterns ...string) map[string]*packages.Package {
	patterns = append(patterns, dot)

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Tests: false,
		Dir:   g.commonFlags.Dir,
	}
	pkgs, err := loadPkgs(cfg, patterns...)
	if err != nil {
		logx.Fatalf("%s", err)
	}

	for _, pat := range patterns {
		if _, ok := pkgs[pat]; !ok {
			logx.Fatalf("no package found with pattern %s", pat)
		}
	}

	primaryPkg := pkgs[dot]
	if g.commonFlags.FileName != "" {
		//only keep the specified file
		var fs []*ast.File
		for _, f := range primaryPkg.Syntax {
			filename := primaryPkg.Fset.File(f.Pos()).Name()
			if filepath.Base(filename) == g.commonFlags.FileName {
				fs = append(fs, f)
				break
			}
		}
		primaryPkg.Syntax = fs
	} else if Contains(g.commonFlags.TypeNames, star) {
		//find the file in which cmdline exists
	end:
		for i, f := range primaryPkg.Syntax {
			for _, cg := range f.Comments {
				for _, c := range cg.List {
					if !findCmdLine(c.Text, g.commonFlags.CmdLine) {
						continue
					}
					filename := filepath.Base(primaryPkg.GoFiles[i])
					g.allInOneFile = filename
					break end
				}
			}
		}
	}

	g.SetPkg(primaryPkg)
	return pkgs
}

func loadPkgs(cfg *packages.Config, patterns ...string) (map[string]*packages.Package, error) {
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*packages.Package, len(patterns))
	for _, pat := range patterns {
		for _, pkg := range pkgs {
			if hasMultiPkgs(pkg) {
				logx.Fatalf("multiple packages found in %s", pkg.Dir)
			}

			if pkg.PkgPath == pat {
				result[pat] = pkg
				break
			}

			isPath := strings.HasPrefix(pat, ".") || strings.HasPrefix(pat, "/")
			if isPath {
				absPat, _ := filepath.Abs(filepath.Join(cfg.Dir, pat))
				if pkg.Dir == absPat {
					_, ok := result[pat]
					if ok {
						logx.Fatalf("multiple packages found in %s", pat)
					}
					result[pat] = pkg
				}
			}
		}
	}

	return result, nil
}

func getGoFile(pkg *packages.Package, typeName string) string {
	for _, obj := range pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}

		_, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		if obj.Name() == typeName {
			pos := pkg.Fset.Position(obj.Pos())
			return filepath.Base(pos.Filename)
		}
	}
	return ""
}

func (g *GeneratorBase) confirmTypes(typeLister TypeLister) {
	typeNames := g.commonFlags.TypeNames
	if g.isTypeSpecified {
		for _, typName := range typeNames {
			gofile := getGoFile(g.pkg, typName)
			if g.commonFlags.FileName == "" {
				g.fileNameMap[typName] = gofile
			} else if g.commonFlags.FileName != gofile {
				logx.Fatalf("type %s is not in the specified file", typName)
			}
		}
	} else {
		g.commonFlags.TypeNames = typeLister.ListTypes()
	}
}

func (g *GeneratorBase) Generate(
	gen interface {
		TypeLister
		DataMaker
	}) map[string][]byte {
	if g.pkg == nil {
		logx.Fatal("primary pkg is nil, may forget to call LoadPackage")
	}

	g.confirmTypes(gen)

	srcMap := make(map[string][]byte)
	var srcList [][]byte
	for _, typName := range g.commonFlags.TypeNames {
		data := gen.MakeData(typName)
		if data == nil {
			continue
		}
		src := g.generateOne(data)
		if len(src) == 0 {
			continue
		}
		if g.commonFlags.Separate {
			srcMap[g.fileName(typName, false)] = src
		} else {
			srcList = append(srcList, src)
		}
	}
	if len(srcList) > 0 {
		src, err := MergeSources(srcList...)
		if err != nil {
			logx.Fatalf("merge sources error: %s", err)
		}
		if len(src) > 0 {
			srcMap[g.fileName("", false)] = src
		}
	}

	return srcMap
}

func (g *GeneratorBase) generateOne(data any) []byte {
	cflags := g.commonFlags
	if cflags.Verbose {
		logx.DebugJSON("template data:\n", data)
	}

	var buff bytes.Buffer
	err := g.tmpl().Execute(&buff, data)
	if err != nil {
		logx.Fatalf("executing template: %s", err)
	}
	src := buff.Bytes()
	if cflags.Verbose {
		logx.Debug("raw source code:\n", string(src))
	}

	if cflags.Raw {
		return src //typically used for debugging
	}

	src, err = FormatSrc(src)
	if err != nil {
		logx.Fatalf("format source: %s", err)
	}
	return src
}

func hasMultiPkgs(pkg *packages.Package) bool {
	for _, e := range pkg.Errors {
		if strings.Contains(e.Msg, "found packages") {
			return true
		}
	}
	return false
}

func findCmdLine(doc string, cmdline string) bool {
	pat := fmt.Sprintf("(?m)^//go:generate.*%s$", regexp.QuoteMeta(cmdline))
	reg := regexp.MustCompile(pat)
	return reg.MatchString(doc)
}

func (g *GeneratorBase) Clean() error {
	if g.commonFlags.Separate {
		return nil
	}
	if g.allInOneFile == "" {
		return nil
	}

	genfile := g.fileName("", false)
	pattern := fmt.Sprintf("*.%s%s*.go", Shoot, g.subCmd)
	glob := filepath.Join(g.commonFlags.Dir, pattern)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, file := range matches {
		if file == genfile {
			continue
		}
		isAIO, err := isAllInOneFile(file)
		if err != nil {
			return err
		}
		if isAIO {
			continue
		}
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}

func isAllInOneFile(file string) (bool, error) {
	line, err := firstLine(file)
	if err != nil {
		return false, err
	}
	pat := "^// Code generated by.*-type=\\*.*DO NOT EDIT."
	reg := regexp.MustCompile(pat)
	return reg.MatchString(line), nil
}

func firstLine(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}
