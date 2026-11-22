//go:build go1.27

package mapper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenericMapperFuncMatchesInstantiatedTypes(t *testing.T) {
	fset := token.NewFileSet()
	typeFile, err := parser.ParseFile(fset, "sqlmapper.go", `package mapper

type Null[T any] struct {
	V T
	Valid bool
}

type SQLMapper struct{}
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	methodFile, err := parser.ParseFile(fset, "sqlmapper_go127.go", `package mapper

func (*SQLMapper) PtrToNullType[T any](value *T) Null[T] {
	if value == nil {
		return Null[T]{}
	}
	return Null[T]{V: *value, Valid: true}
}

type Fields struct {
	Source *int
	Destination Null[int]
	WrongDestination Null[string]
}
`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{GoVersion: "go1.27"}
	pkg, err := conf.Check("example.com/mapper", fset, []*ast.File{typeFile, methodFile}, info)
	if err != nil {
		t.Fatal(err)
	}

	g := New()
	g.mapperpkg = &packages.Package{
		Name:      pkg.Name(),
		PkgPath:   pkg.Path(),
		Syntax:    []*ast.File{typeFile, methodFile},
		Types:     pkg,
		TypesInfo: info,
	}
	g.parseMapper("SQLMapper")
	if len(g.mappingFuncList) != 1 {
		t.Fatalf("mapping funcs = %d, want 1", len(g.mappingFuncList))
	}
	fn := g.mappingFuncList[0]

	fields := pkg.Scope().Lookup("Fields").Type().Underlying().(*types.Struct)
	source := fields.Field(0).Type()
	destination := fields.Field(1).Type()
	wrongDestination := fields.Field(2).Type()
	if !mapperFuncMatches(fn, source, destination) {
		t.Fatal("generic mapper method should match *int -> Null[int]")
	}
	if mapperFuncMatches(fn, source, wrongDestination) {
		t.Fatal("generic mapper method must not match *int -> Null[string]")
	}
}
