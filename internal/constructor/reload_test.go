package constructor

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateTwiceWithExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example\n\ngo 1.24.0\n")
	writeTestFile(t, filepath.Join(dir, "p.go"), `package p

type User struct {
	id string
}

type Order struct {
	id string
}
`)

	first := generateNew(t, dir)
	merged, ok := first["p.shootnew.go"]
	if !ok {
		t.Fatalf("first generate files: %v", keys(first))
	}
	writeTestFile(t, filepath.Join(dir, "p.shootnew.go"), string(merged))
	writeTestFile(t, filepath.Join(dir, "use.go"), `package p

func keepBoth() (*User, *Order) {
	return NewUser("a"), NewOrder("b")
}
`)

	second := generateNew(t, dir)
	if _, ok := second["p.shootnew.go"]; !ok {
		t.Fatalf("second generate files: %v", keys(second))
	}
}

func generateNew(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	oldArgs := os.Args
	oldCmd := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCmd
	}()

	os.Args = []string{"shoot", "new", "-getset", "-file=p.go", "-version=test", dir}
	flag.CommandLine = flag.NewFlagSet("shoot", flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}

	g := New()
	g.ParseFlags()
	g.LoadPackage()
	return g.Generate(g)
}

func writeTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
