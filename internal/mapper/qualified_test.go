package mapper

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

// -alias applies to the destination package only. time.Time must stay time.Time.
// With no alias, a named type in the current package must not be written as p.ID.
func TestQualifiedTypeName(t *testing.T) {
	pkg := typecheck(t, `
package p

import "time"

type ID int

type Src struct {
	When time.Time
	Code ID
}
`)

	g := New()
	g.SetPkg(&packages.Package{Types: pkg, PkgPath: pkg.Path()})
	g.destPkg = &packages.Package{Name: "dest", PkgPath: "example.com/dest"}

	t.Run("alias", func(t *testing.T) {
		g.flags = &Flags{alias: "dest"}
		got := g.qualifiedTypeName(fieldType(t, pkg, "When"))
		if got != "time.Time" {
			t.Fatalf("got %s, want time.Time", got)
		}
	})
	t.Run("current package", func(t *testing.T) {
		g.flags = &Flags{}
		got := g.qualifiedTypeName(fieldType(t, pkg, "Code"))
		if got != "ID" {
			t.Fatalf("got %s, want ID", got)
		}
	})
}

func typecheck(t *testing.T, src string) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("p", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

func fieldType(t *testing.T, pkg *types.Package, name string) types.Type {
	t.Helper()
	obj := pkg.Scope().Lookup("Src")
	st := obj.Type().Underlying().(*types.Struct)
	for i := 0; i < st.NumFields(); i++ {
		if st.Field(i).Name() == name {
			return st.Field(i).Type()
		}
	}
	t.Fatalf("field %s not found", name)
	return nil
}
