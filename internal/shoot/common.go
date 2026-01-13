package shoot

import (
	"go/types"
	"path/filepath"
	"strings"
)

func Contains[T comparable](slice []T, val T) bool {
	for _, x := range slice {
		if x == val {
			return true
		}
	}
	return false
}

func FixPath(path string) string {
	if path == "" {
		return dot
	}
	if strings.HasPrefix(path, dot) || filepath.IsAbs(path) {
		return path
	}
	return "./" + path
}

func TypeEquals(x, y types.Type) bool {
	if x == nil || y == nil {
		return x == y
	}
	nopkg := false
	qf := func(p *types.Package) string {
		if p == nil {
			nopkg = true
			return ""
		}
		return p.Path()
	}
	xname := types.TypeString(x, qf)
	yname := types.TypeString(y, qf)
	if nopkg {
		return false
	}
	return xname == yname
}
