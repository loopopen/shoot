package mapper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"testing"

	"github.com/loopopen/shoot/internal/shoot"
	"golang.org/x/tools/go/packages"
)

func TestFindImportForSelectorUsesDeclaredPackageName(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", `package p

import (
	"example.com/api/v2"
	m "example.com/other/v2"
	"example.com/lib/notmapper"
)

type Src struct {
	api.Mapper
	m.Mapper
	custom.Mapper
}
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	local := types.NewPackage("p", "p")
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
	selectors := embeddedSelectors(t, file)
	usePkg(info, local, selectors["api"], "example.com/api/v2", "api")
	usePkg(info, local, selectors["m"], "example.com/other/v2", "m")
	usePkg(info, local, selectors["custom"], "example.com/lib/notmapper", "custom")

	got := importPath(t, findImportForSelector(file, selectors["api"], info))
	if got != "example.com/api/v2" {
		t.Fatalf("v2 import path = %s", got)
	}
	got = importPath(t, findImportForSelector(file, selectors["m"], info))
	if got != "example.com/other/v2" {
		t.Fatalf("aliased v2 import path = %s", got)
	}
	got = importPath(t, findImportForSelector(file, selectors["custom"], info))
	if got != "example.com/lib/notmapper" {
		t.Fatalf("renamed package import path = %s", got)
	}
}

func embeddedSelectors(t *testing.T, file *ast.File) map[string]*ast.SelectorExpr {
	t.Helper()
	out := map[string]*ast.SelectorExpr{}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if ok {
			out[ident.Name] = sel
		}
		return true
	})
	return out
}

func usePkg(info *types.Info, local *types.Package, sel *ast.SelectorExpr, path, name string) {
	ident := sel.X.(*ast.Ident)
	imported := types.NewPackage(path, name)
	info.Uses[ident] = types.NewPkgName(ident.Pos(), local, ident.Name, imported)
}

func importPath(t *testing.T, imp *ast.ImportSpec) string {
	t.Helper()
	if imp == nil || imp.Path == nil {
		t.Fatal("import not found")
	}
	path, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNamedScalarTypesUseConversionInsteadOfRecursiveMapping(t *testing.T) {
	srcPkg := types.NewPackage("example.test/src", "src")
	destPkg := types.NewPackage("example.test/dest", "dest")
	srcLabel := types.NewNamed(types.NewTypeName(token.NoPos, srcPkg, "Label", nil), types.Typ[types.Int], nil)
	destLabel := types.NewNamed(types.NewTypeName(token.NoPos, destPkg, "Label", nil), types.Typ[types.Int32], nil)

	srcField := &Field{Name: "Label", typ: srcLabel}
	destField := &Field{Name: "Label", typ: destLabel}
	g := New()
	g.SetPkg(&packages.Package{Types: srcPkg, PkgPath: srcPkg.Path()})
	g.destPkg = &packages.Package{Types: destPkg, Name: destPkg.Name(), PkgPath: destPkg.Path()}
	g.flags = &Flags{}
	g.exportedFields = []*Field{srcField}
	g.destExportedFields = []*Field{destField}
	g.writeSrcSet = shoot.MakeSet[string]()
	g.writeDestSet = shoot.MakeSet[string]()

	g.makeTypeMismatch()
	if srcField.Target != nil || destField.Target != nil || srcField.CanMap || destField.CanMap {
		t.Fatal("named scalar types must not be marked for recursive mapping")
	}

	g.makeTypeMatch()
	if !srcField.IsConv || !destField.IsConv {
		t.Fatal("convertible named scalar types must use explicit conversion")
	}
	if destField.Type != "dest.Label" || srcField.Type != "Label" {
		t.Fatalf("conversion types = (%q, %q), want (dest.Label, Label)", destField.Type, srcField.Type)
	}
}
