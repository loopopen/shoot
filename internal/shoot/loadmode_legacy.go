//go:build !go1.27

package shoot

import "golang.org/x/tools/go/packages"

func packageLoadMode() packages.LoadMode {
	return basePackageLoadMode()
}
