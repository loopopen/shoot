//go:build !go1.27

package main

import (
	"strings"
	"testing"

	"github.com/loopopen/shoot/internal/constructor"
)

func TestConstructorNestsPromotedFieldsBeforeGo127(t *testing.T) {
	generated := generatePromotedConstructor(t)
	if !strings.Contains(generated, "PaymentBase: PaymentBase[uint]{") {
		t.Fatalf("legacy constructor should initialize the embedded struct:\n%s", generated)
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
