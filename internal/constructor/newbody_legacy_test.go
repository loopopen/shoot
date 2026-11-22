//go:build !go1.27

package constructor

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestNewBodyNestsPromotedFieldsBeforeGo127(t *testing.T) {
	fields, nameMap := promotedFieldFixture()
	want := "Base: Base[uint]{\n" +
		"id: id,\n" +
		"},\n" +
		"payNo: payNo,\n"
	if got := newBody(fields, nameMap, true); got != want {
		t.Fatalf("newBody() = %q, want %q", got, want)
	}
	if supportsPromotedFieldLiterals(&packages.Package{Module: &packages.Module{GoVersion: "1.27.0"}}) {
		t.Fatal("legacy generator enabled Go 1.27 promoted-field literals")
	}
}
