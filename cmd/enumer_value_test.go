package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/loopopen/shoot/internal/enumer"
)

func TestUint64EnumAboveMaxInt64Compiles(t *testing.T) {
	const filename = "enum_wide.shootenum.wide.go"
	srcMap := generate(Golden{
		cmd: "shoot enum -type=Wide",
		names: []string{
			filename,
		},
	}, enumer.New(), "./test/enumer")
	generated, ok := srcMap[filename]
	if !ok {
		t.Fatalf("generated files: %v", srcMap)
	}

	origin, err := os.ReadFile("./test/enumer/enum_wide.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := typecheckEnumer("enum_wide.go", origin, filename, generated); err != nil {
		t.Fatal(err)
	}
}

func TestSignedEnumSortsNegativesFirst(t *testing.T) {
	const filename = "enum_signed.shootenum.rank.go"
	srcMap := generate(Golden{
		cmd: "shoot enum -type=Rank",
		names: []string{
			filename,
		},
	}, enumer.New(), "./test/enumer")
	generated, ok := srcMap[filename]
	if !ok {
		t.Fatalf("generated files: %v", srcMap)
	}

	const want = "[]Rank{RankNeg, RankZero, RankPos}"
	if !strings.Contains(string(generated), want) {
		t.Fatalf("signed values should sort from negative to positive, got:\n%s", generated)
	}
}

func TestEnumCaseUsesOneStringRepresentation(t *testing.T) {
	const filename = "enum_case.shootenum.httpstatus.go"
	srcMap := generate(Golden{
		cmd: "shoot enum -json -text -sql -case=snake -type=HTTPStatus",
		names: []string{
			filename,
		},
	}, enumer.New(), "./test/enumer")
	generated, ok := srcMap[filename]
	if !ok {
		t.Fatalf("generated files: %v", srcMap)
	}

	source := string(generated)
	for _, want := range []string{
		`[]string{"ok", "not_found", "http_server_error"}`,
		`"not_found":`,
		`return json.Marshal(h.String())`,
		`return []byte(h.String()), nil`,
		`return h.String(), nil`,
	} {
		if !strings.Contains(source, want) {
			t.Errorf("generated enum missing %q:\n%s", want, source)
		}
	}
	if got := strings.Count(source, "shoot.ParseEnum[HTTPStatus]"); got != 3 {
		t.Errorf("JSON, Text, and SQL should all parse the same enum strings; ParseEnum calls = %d, want 3", got)
	}
}

func typecheckEnumer(originName string, origin []byte, generatedName string, generated []byte) error {
	fset := token.NewFileSet()
	var files []*ast.File
	for name, src := range map[string][]byte{
		originName:    origin,
		generatedName: generated,
	} {
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			return err
		}
		files = append(files, file)
	}
	conf := types.Config{Importer: importer.Default()}
	_, err := conf.Check("enumer", fset, files, nil)
	return err
}
