package shoot

import "golang.org/x/tools/go/packages"

func basePackageLoadMode() packages.LoadMode {
	return packages.NeedName |
		packages.NeedFiles |
		packages.NeedModule |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo
}
