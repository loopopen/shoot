package shoot

import (
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestLoadPkgsStopsOnPackageErrors(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		dir := writeModule(t, "package bad\n\nvar x = undefined\n")
		_, err := loadDir(t, dir)
		if err == nil || !strings.Contains(err.Error(), "undefined") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		dir := writeModule(t, "package bad\n\nfunc (\n")
		_, err := loadDir(t, dir)
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("multiple packages", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "go.mod"), "module example\n\ngo 1.24.0\n")
		writeFile(t, filepath.Join(dir, "a.go"), "package a\n")
		writeFile(t, filepath.Join(dir, "b.go"), "package b\n")
		_, err := loadDir(t, dir)
		if err == nil || !strings.Contains(err.Error(), "multiple packages found") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("valid package", func(t *testing.T) {
		dir := writeModule(t, "package ok\n\nvar x = 1\n")
		pkgs, err := loadDir(t, dir)
		if err != nil {
			t.Fatal(err)
		}
		if pkgs["."] == nil {
			t.Fatalf("missing package: %v", pkgs)
		}
	})
}

func writeModule(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(dir, "p.go"), src)
	return dir
}

func loadDir(t *testing.T, dir string) (map[string]*packages.Package, error) {
	t.Helper()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir: dir,
	}
	return loadPkgs(cfg, ".")
}
