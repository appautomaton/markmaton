package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func RepoRoot(t testing.TB) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repo root: runtime caller unavailable")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func ReadFixture(t testing.TB, name string) string {
	t.Helper()

	path := filepath.Join(RepoRoot(t), "testdata", "fixtures", filepath.FromSlash(name))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return string(data)
}

func ReadGolden(t testing.TB, name string) string {
	t.Helper()

	path := filepath.Join(RepoRoot(t), "testdata", "golden", filepath.FromSlash(name))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %q: %v", name, err)
	}
	return string(data)
}
