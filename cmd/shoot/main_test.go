package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotedownSrcWritesIntoDir(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "model")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "user.shootnew.go")
	notedownSrc(dir, outPath, []byte("package model\n"))

	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "user.shootnew.go")); !os.IsNotExist(err) {
		t.Fatalf("wrote into the parent directory: %v", err)
	}
}
