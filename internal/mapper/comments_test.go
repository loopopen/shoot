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
	if got, want := updates[targetFset.Position(targetID.Pos()).Offset].text, `// source identifier`; got != want {
		t.Fatalf(`comment = %q, want %q`, got, want)
	}
	if got, want := updates[targetFset.Position(targetOrderingTime.Pos()).Offset].text, `// order time`; got != want {
		t.Fatalf(`tag-matched comment = %q, want %q`, got, want)
	}
	if _, ok := updates[targetFset.Position(targetKeep.Pos()).Offset]; ok {
		t.Fatal(`field with a comment was overwritten`)
	}
}

func TestCollectCommentUpdatesOff(t *testing.T) {
	g := commentFixture(t)
	g.flags.comment = CommentOff
	g.collectCommentUpdates()
	if len(g.commentUpdates) != 0 {
		t.Fatalf(`comment updates = %d, want 0`, len(g.commentUpdates))
	}
}

func TestCollectCommentUpdatesForce(t *testing.T) {
	g := commentFixture(t)
	g.flags.comment = CommentForce
	g.collectCommentUpdates()

	targetKeep := g.exportedFields[2].node
	filename := g.Pkg().Fset.Position(g.exportedFields[0].node.Pos()).Filename
	updates := g.commentUpdates[filename]
	if len(updates) != 3 {
		t.Fatalf(`comment updates = %d, want 3`, len(updates))
	}
	keep := updates[g.Pkg().Fset.Position(targetKeep.Pos()).Offset]
	if keep.text != `// source documentation` {
		t.Fatalf(`comment = %q, want %q`, keep.text, `// source documentation`)
	}
	if !keep.replace {
		t.Fatal(`existing line comment was not replaced`)
	}
	from := g.Pkg().Fset.Position(targetKeep.Comment.Pos()).Offset
	to := g.Pkg().Fset.Position(targetKeep.Comment.End()).Offset
	if keep.from != from || keep.to != to {
		t.Fatalf(`replacement = [%d,%d), want [%d,%d)`, keep.from, keep.to, from, to)
	}
}

func TestCollectCommentUpdatesForceClearsEmpty(t *testing.T) {
	targetSrc := []byte(`package target

type Target struct {
	Noted string // keep me
	Plain string
}
`)
	sourceSrc := []byte(`package source

type Source struct {
	Noted string
	Plain string
}
`)
	targetFset, targetFile := parseCommentTestFile(t, targetSrc)
	_, sourceFile := parseCommentTestFile(t, sourceSrc)
	noted := structField(t, targetFile, `Target`, `Noted`)
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
		flags:         &Flags{comment: CommentForce},
		exportedFields: []*Field{
			{Name: `Noted`, node: noted},
			{Name: `Plain`, node: structField(t, targetFile, `Target`, `Plain`)},
		},
		destExportedFields: []*Field{
			{Name: `Noted`, node: structField(t, sourceFile, `Source`, `Noted`)},
			{Name: `Plain`, node: structField(t, sourceFile, `Source`, `Plain`)},
		},
	}
	g.SetPkg(&packages.Package{Fset: targetFset})
	g.collectCommentUpdates()

	filename := targetFset.Position(noted.Pos()).Filename
	updates := g.commentUpdates[filename]
	if len(updates) != 1 {
		t.Fatalf(`comment updates = %d, want 1`, len(updates))
	}
	edit := updates[targetFset.Position(noted.Pos()).Offset]
	if edit.text != `` || !edit.replace {
		t.Fatalf(`clear edit = %+v, want an empty replacement`, edit)
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
	updates := map[int]commentEdit{
		fset.Position(id.Pos()).Offset:   {text: `// identifier`},
		fset.Position(name.Pos()).Offset: {text: `// display name`},
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

func TestApplyCommentUpdatesForce(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, `target.go`)
	src := []byte(`package target

type Target struct {
	// old identifier
	ID string
	Name string // old name
}
`)
	if err := os.WriteFile(filename, src, 0o644); err != nil {
		t.Fatal(err)
	}

	fset, file := parseCommentTestFile(t, src)
	id := structField(t, file, `Target`, `ID`)
	name := structField(t, file, `Target`, `Name`)
	idRemove, _, _, _ := commentReplacement(fset, id)
	_, nameReplace, nameFrom, nameTo := commentReplacement(fset, name)
	if len(idRemove) != 1 || !nameReplace {
		t.Fatalf(`replacement spans = doc %d, line %v`, len(idRemove), nameReplace)
	}
	updates := map[int]commentEdit{
		fset.Position(id.Pos()).Offset: {
			text:   `// identifier`,
			remove: idRemove,
		},
		fset.Position(name.Pos()).Offset: {
			text:    `// display name`,
			replace: true,
			from:    nameFrom,
			to:      nameTo,
		},
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

func TestApplyCommentUpdatesForceClears(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, `target.go`)
	src := []byte(`package target

type Target struct {
	// old identifier
	ID string
	Name string // old name
	Plain string
}
`)
	if err := os.WriteFile(filename, src, 0o644); err != nil {
		t.Fatal(err)
	}

	fset, file := parseCommentTestFile(t, src)
	id := structField(t, file, `Target`, `ID`)
	name := structField(t, file, `Target`, `Name`)
	idRemove, _, _, _ := commentReplacement(fset, id)
	_, nameReplace, nameFrom, nameTo := commentReplacement(fset, name)
	updates := map[int]commentEdit{
		fset.Position(id.Pos()).Offset:   {remove: idRemove},
		fset.Position(name.Pos()).Offset: {replace: nameReplace, from: nameFrom, to: nameTo},
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
	ID    string
	Name  string
	Plain string
}
`
	if string(got) != want {
		t.Fatalf(`updated source =
%s
want:
%s`, got, want)
	}
}

func TestCommentMode(t *testing.T) {
	var comment Comment
	if comment.String() != `on` {
		t.Fatalf(`default = %q, want on`, comment.String())
	}
	if err := comment.Set(`force`); err != nil {
		t.Fatal(err)
	}
	if comment != CommentForce {
		t.Fatalf(`comment = %q, want force`, comment)
	}
	if err := comment.Set(`sideways`); err == nil {
		t.Fatal(`invalid comment mode was accepted`)
	}
}

func commentFixture(t *testing.T) *Generator {
	t.Helper()
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
	g := &Generator{
		GeneratorBase: shoot.NewGeneratorBase(SubCmd, tmplTxt),
		flags:         &Flags{},
		exportedFields: []*Field{
			{Name: `ID`, node: structField(t, targetFile, `Target`, `ID`)},
			{Name: `OrderingTime`, node: structField(t, targetFile, `Target`, `OrderingTime`)},
			{Name: `Keep`, node: structField(t, targetFile, `Target`, `Keep`)},
		},
		destExportedFields: []*Field{
			{Name: `ID`, node: structField(t, sourceFile, `Source`, `ID`)},
			{Name: `OrderTime`, node: structField(t, sourceFile, `Source`, `OrderTime`)},
			{Name: `Keep`, node: structField(t, sourceFile, `Source`, `Keep`)},
		},
		srcTagMap: map[string]string{`OrderingTime`: `OrderTime`},
	}
	g.SetPkg(&packages.Package{Fset: targetFset})
	return g
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
