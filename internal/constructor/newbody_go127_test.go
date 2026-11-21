//go:build go1.27

package constructor

import "testing"

import "golang.org/x/tools/go/packages"

type promotedBaseGo127[T any] struct {
	id T
}

type paymentOrderGo127 struct {
	promotedBaseGo127[uint]
	payNo string
}

func TestNewBodyAssignsPromotedFieldsDirectlyInGo127(t *testing.T) {
	fields, nameMap := promotedFieldFixture()
	want := "id: id,\n" +
		"payNo: payNo,\n"
	if got := newBody(fields, nameMap, true); got != want {
		t.Fatalf("newBody() = %q, want %q", got, want)
	}

	order := paymentOrderGo127{id: 1, payNo: "PAY-1"}
	if order.id != 1 || order.payNo != "PAY-1" {
		t.Fatalf("direct promoted-field initialization failed: %+v", order)
	}
}

func TestNewBodyKeepsPointerEmbeddedFieldsNestedInGo127(t *testing.T) {
	fields, nameMap := promotedFieldFixture()
	fields[0].isPtr = true
	want := "Base: &Base[uint]{\n" +
		"id: id,\n" +
		"},\n" +
		"payNo: payNo,\n"
	if got := newBody(fields, nameMap, true); got != want {
		t.Fatalf("newBody() = %q, want %q", got, want)
	}
}

func TestSupportsPromotedFieldLiteralsRequiresGo127Module(t *testing.T) {
	for goVersion, want := range map[string]bool{
		"":       false,
		"1.24.0": false,
		"1.26.8": false,
		"1.27.0": true,
		"1.28.0": true,
	} {
		pkg := &packages.Package{Module: &packages.Module{GoVersion: goVersion}}
		if got := supportsPromotedFieldLiterals(pkg); got != want {
			t.Errorf("supportsPromotedFieldLiterals(go %s) = %v, want %v", goVersion, got, want)
		}
	}
}
