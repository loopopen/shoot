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
