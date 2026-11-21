//go:build go1.27

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loopopen/shoot/internal/constructor"
)

func TestConstructorKeepsLegacyLiteralForGo124Module(t *testing.T) {
	generated := generatePromotedConstructor(t)
	if !strings.Contains(generated, "PaymentBase: PaymentBase[uint]{") {
		t.Fatalf("Go 1.24 module should retain the legacy embedded literal:\n%s", generated)
	}
}

func TestConstructorAssignsPromotedFieldsDirectlyInGo127Module(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n\ngo 1.27.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	const source = `package example

type Base[T any] struct {
	id T
}

type PaymentOrder struct {
	Base[uint]
	payNo string
}
`
	if err := os.WriteFile(filepath.Join(dir, "payment.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	const filename = "payment.shootnew.paymentorder.go"
	srcMap := generate(Golden{
		cmd:   "shoot new -type=PaymentOrder",
		names: []string{filename},
	}, constructor.New(), dir)
	generated, ok := srcMap[filename]
	if !ok {
		t.Fatalf("generated files: %v", srcMap)
	}
	result := string(generated)
	if !strings.Contains(result, "return &PaymentOrder{\n\t\tid:    id,\n\t\tpayNo: payNo,") {
		t.Fatalf("Go 1.27 constructor should initialize promoted fields directly:\n%s", result)
	}
	if strings.Contains(result, "Base: Base[uint]{") {
		t.Fatalf("Go 1.27 constructor unexpectedly nested the embedded struct:\n%s", result)
	}

	fset := token.NewFileSet()
	originFile, err := parser.ParseFile(fset, "payment.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	generatedFile, err := parser.ParseFile(fset, filename, generated, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{GoVersion: "go1.27"}
	if _, err := conf.Check("example", fset, []*ast.File{originFile, generatedFile}, nil); err != nil {
		t.Fatal(err)
	}
}

func generatePromotedConstructor(t *testing.T) string {
	t.Helper()
	const filename = "new_promoted.shootnew.paymentorder.go"
	srcMap := generate(Golden{
		cmd:   "shoot new -type=PaymentOrder",
		names: []string{filename},
	}, constructor.New(), "./test/ctor")
	generated, ok := srcMap[filename]
	if !ok {
		t.Fatalf("generated files: %v", srcMap)
	}
	return string(generated)
}
