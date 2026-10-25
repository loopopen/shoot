package mapper

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/loopopen/shoot/internal/shoot"
	"golang.org/x/tools/go/packages"
)

func TestNamedScalarTypesUseConversionInsteadOfRecursiveMapping(t *testing.T) {
	srcPkg := types.NewPackage("example.test/src", "src")
	destPkg := types.NewPackage("example.test/dest", "dest")
	srcLabel := types.NewNamed(types.NewTypeName(token.NoPos, srcPkg, "Label", nil), types.Typ[types.Int], nil)
	destLabel := types.NewNamed(types.NewTypeName(token.NoPos, destPkg, "Label", nil), types.Typ[types.Int32], nil)

	srcField := &Field{Name: "Label", typ: srcLabel}
	destField := &Field{Name: "Label", typ: destLabel}
	g := New()
	g.SetPkg(&packages.Package{Types: srcPkg, PkgPath: srcPkg.Path()})
	g.destPkg = &packages.Package{Types: destPkg, Name: destPkg.Name(), PkgPath: destPkg.Path()}
	g.flags = &Flags{}
	g.exportedFields = []*Field{srcField}
	g.destExportedFields = []*Field{destField}
	g.writeSrcSet = shoot.MakeSet[string]()
	g.writeDestSet = shoot.MakeSet[string]()

	g.makeTypeMismatch()
	if srcField.Target != nil || destField.Target != nil || srcField.CanMap || destField.CanMap {
		t.Fatal("named scalar types must not be marked for recursive mapping")
	}

	g.makeTypeMatch()
	if !srcField.IsConv || !destField.IsConv {
		t.Fatal("convertible named scalar types must use explicit conversion")
	}
	if destField.Type != "dest.Label" || srcField.Type != "Label" {
		t.Fatalf("conversion types = (%q, %q), want (dest.Label, Label)", destField.Type, srcField.Type)
	}
}
