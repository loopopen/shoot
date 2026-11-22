//go:build go1.27

package shoot

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGo127PackageLoadModeIncludesDependencies(t *testing.T) {
	if packageLoadMode()&packages.NeedDeps == 0 {
		t.Fatal("Go 1.27 package load mode must include dependencies")
	}
}
