package mapper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/loopopen/shoot/internal/shoot"
	"golang.org/x/tools/go/packages"
)

func TestCollectCommentUpdates(t *testing.T) {
	targetSrc := []byte(`package target

type Target struct {
	ID           string
	OrderingTime string
	Keep         string // target documentation
}
`)
	sourceSrc := []byte(`package source

type Source struct {
	// source identifier
	ID string
	// order time
	OrderTime string
	// source documentation
	Keep string
}
`)

	targetFset, targetFile := parseCommentTestFile(t, targetSrc)
	_, sourceFile := parseCommentTestFile(t, sourceSrc)
	targetID := structField(t, targetFile, `Target`, `ID`)
	targetOrderingTime := structField(t, targetFile, `Target`, `OrderingTime`)
	targetKeep := structField(t, targetFile, `Target`, `Keep`)
	sourceID := structField(t, sourceFile, `Source`, `ID`)
	sourceOrderTime := structField(t, sourceFile, `Source`, `OrderTime`)
	sourceKeep := structField(t, sourceFile, `Source`, `Keep`)

	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
		flags:         &Flags{},
		exportedFields: []*Field{
			{Name: `ID`, node: targetID},
			{Name: `OrderingTime`, node: targetOrderingTime},
			{Name: `Keep`, node: targetKeep},
		},
		destExportedFields: []*Field{
			{Name: `ID`, node: sourceID},
			{Name: `OrderTime`, node: sourceOrderTime},
			{Name: `Keep`, node: sourceKeep},
		},
		srcTagMap: map[string]string{`OrderingTime`: `OrderTime`},
	}
	g.SetPkg(&packages.Package{Fset: targetFset})

	g.collectCommentUpdates()

	filename := targetFset.Position(targetID.Pos()).Filename
	updates := g.commentUpdates[filename]
	if len(updates) != 2 {
		t.Fatalf(`comment updates = %d, want 2`, len(updates))
	}
	if got, want := updates[targetFset.Position(targetID.Pos()).Offset], `// source identifier`; got != want {
		t.Fatalf(`comment = %q, want %q`, got, want)
	}
	if got, want := updates[targetFset.Position(targetOrderingTime.Pos()).Offset], `// order time`; got != want {
		t.Fatalf(`tag-matched comment = %q, want %q`, got, want)
	}
}

func TestApplyCommentUpdates(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, `target.go`)
	src := []byte(`package target

type Target struct {
	ID string
	Name string
}
`)
	if err := os.WriteFile(filename, src, 0o644); err != nil {
		t.Fatal(err)
	}

	fset, file := parseCommentTestFile(t, src)
	id := structField(t, file, `Target`, `ID`)
	name := structField(t, file, `Target`, `Name`)
	updates := map[int]string{
		fset.Position(id.Pos()).Offset:   `// identifier`,
		fset.Position(name.Pos()).Offset: `// display name`,
	}
	if err := applyCommentUpdates(filename, updates); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	want := `package target

type Target struct {
	ID   string // identifier
	Name string // display name
}
`
	if string(got) != want {
		t.Fatalf(`updated source =
%s
want:
%s`, got, want)
	}
}

func parseCommentTestFile(t *testing.T, src []byte) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, `target.go`, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return fset, file
}

func structField(t *testing.T, file *ast.File, typeName, fieldName string) *ast.Field {
	t.Helper()
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if name.Name == fieldName {
						return field
					}
				}
			}
		}
	}
	t.Fatalf(`field %s.%s not found`, typeName, fieldName)
	return nil
}
