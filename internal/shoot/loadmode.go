package shoot

import "golang.org/x/tools/go/packages"

func basePackageLoadMode() packages.LoadMode {
	return packages.NeedName |
		packages.NeedFiles |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo
}
