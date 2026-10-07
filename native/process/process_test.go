package process

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dyike/keel/native"
)

func TestInvalidArguments(t *testing.T) {
	if _, err := ForegroundPID(nil); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := Directory(0); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal(err)
	}
}
func TestCurrentProcessDirectory(t *testing.T) {
	dir, err := Directory(os.Getpid())
	if errors.Is(err, native.ErrUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, _ = filepath.EvalSymlinks(dir)
	expected, _ = filepath.EvalSymlinks(expected)
	if dir != expected {
		t.Fatalf("got %q want %q", dir, expected)
	}
	if _, err := Directory(2147483647); err == nil {
		t.Fatal("missing process succeeded")
	}
}
func TestNonTerminal(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("unsupported")
	}
	f, err := os.CreateTemp(t.TempDir(), "regular")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := ForegroundPID(f); err == nil {
		t.Fatal("regular file accepted as terminal")
	}
}
