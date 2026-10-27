package constructor

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// TestTypeParamConstraintAlignment shows that a constraint which is not a plain
// identifier is dropped. constraints.Ordered is a selector, and int | string is
// a union; both miss the *ast.Ident branch. The remaining constraints then line
// up with the wrong type parameter names.
func TestTypeParamConstraintAlignment(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example\n\ngo 1.24.0\n")
	writeTestFile(t, filepath.Join(dir, "p.go"), `package p

type Num[T int | string, U any] struct {
	v T
	n U
}

type Box[T interface{ ~int | ~string }, U any] struct {
	v T
	n U
}
`)

	g := loadConstructor(t, dir)
	assertTypeParams(t, g, "Num", "T int | string, U any", "T, U")
	assertTypeParams(t, g, "Box", "T interface{ ~int | ~string }, U any", "T, U")
}

func assertTypeParams(t *testing.T, g *Generator, typeName, list, names string) {
	t.Helper()
	data, _ := g.MakeData(typeName)
	td := data.(*TmplData)
	if td.TypeParamList != list || td.TypeParamNameList != names {
		t.Errorf("%s type params:\nTypeParamList = %q\nTypeParamNameList = %q\nwant list %q\nwant names %q",
			typeName, td.TypeParamList, td.TypeParamNameList, list, names)
	}
}

func loadConstructor(t *testing.T, dir string) *Generator {
	t.Helper()
	oldArgs := os.Args
	oldCmd := flag.CommandLine
	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCmd
	})

	os.Args = []string{"shoot", "new", "-file=p.go", "-version=test", dir}
	flag.CommandLine = flag.NewFlagSet("shoot", flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}

	g := New()
	g.ParseFlags()
	g.LoadPackage()
	return g
}
