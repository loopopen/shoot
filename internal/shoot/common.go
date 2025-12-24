package shoot

import (
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
