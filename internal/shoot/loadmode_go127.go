//go:build go1.27

package shoot

import "golang.org/x/tools/go/packages"

func packageLoadMode() packages.LoadMode {
	// The x/tools v0.41 used by the Go 1.24-compatible shoot module cannot
	// read Go 1.27 export data. Loading dependency types from source keeps
	// that build usable in Go 1.27 modules.
	return basePackageLoadMode() | packages.NeedDeps
}
