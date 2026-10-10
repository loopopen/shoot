package constructor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

// BoxSetter has two type parameters and does not match embedded Box[int].
// AssignableToIface returns nil, so the setter branch must not call Underlying.
func TestEmbeddedSetterNilIfaceDoesNotPanic(t *testing.T) {
	pkg := typecheck(t, `
package p

type Box[T any] struct {
	v T
}

type BoxSetter[T any, U any] interface {
	SetV(T)
}

type Holder struct {
	Box[int]
}
`)
	g := newGetSetGenerator(pkg, "Box", embeddedType(t, pkg, "Holder"))

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("generator panicked: %v", rec)
		}
	}()
	g.makeGetSet()
}

// Embed does not implement EmbedSetter. Its methods must not be collected.
func TestEmbeddedSetterIgnoresUnimplementedIface(t *testing.T) {
	pkg := typecheck(t, `
package p

type Embed struct {
	n int
}

type EmbedSetter interface {
	SetN(int)
}

type Wrapper struct {
	Embed
}
`)
	g := newGetSetGenerator(pkg, "Embed", embeddedType(t, pkg, "Wrapper"))
	g.makeGetSet()

	for _, m := range g.getsetMethods {
		if m.Name == "SetN" {
			t.Fatal("collected SetN although Embed does not implement EmbedSetter")
		}
	}
}

// typecheck parses src and returns its type-checked package.
func typecheck(t *testing.T, src string) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := new(types.Config).Check("p", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

// embeddedType returns the type of the first embedded field in holder.
func embeddedType(t *testing.T, pkg *types.Package, holder string) types.Type {
	t.Helper()
	obj := pkg.Scope().Lookup(holder)
	named, ok := obj.Type().(*types.Named)
	if !ok {
		t.Fatalf("%s is not a named type", holder)
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("%s is not a struct", holder)
	}
	field := st.Field(0)
	if !field.Embedded() {
		t.Fatalf("%s.%s is not embedded", holder, field.Name())
	}
	return field.Type()
}

// newGetSetGenerator builds a generator whose only field is the embedded field.
// FindGetterSetterIfac looks up field+"Setter" in pkg.
func newGetSetGenerator(pkg *types.Package, field string, typ types.Type) *Generator {
	g := New()
	g.SetPkg(&packages.Package{
		Types:   pkg,
		PkgPath: pkg.Path(),
	})
	g.getter = true
	g.setter = true
	g.flags = &Flags{getset: true}
	g.data = NewTmplData("test", "test")
	g.fields = []*Field{{
		name:      field,
		isEmbeded: true,
		typ:       typ,
	}}
	return g
}
