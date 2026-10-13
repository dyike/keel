package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTidyFailurePreservesDestination(t *testing.T) {
	// An empty cache and disabled proxy deterministically fail tidy after valid
	// checkout validation, with no compiler or network dependency.
	t.Setenv("GOMODCACHE", t.TempDir())
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GONOPROXY", "none")
	t.Setenv("GOTOOLCHAIN", "local")
	for _, existing := range []bool{false, true} {
		name := "missing"
		if existing {
			name = "existing"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "app")
			if existing {
				if err := os.Mkdir(target, 0755); err != nil {
					t.Fatal(err)
				}
			}
			c, out := newCLI(root)
			if c.main([]string{"new", "app", "-replace", testCheckout(t)}) == 0 {
				t.Fatal("expected tidy failure")
			}
			if !strings.Contains(out.String(), "tidy") {
				t.Fatalf("did not reach tidy: %s", out.String())
			}
			entries, err := os.ReadDir(target)
			if existing {
				if err != nil || len(entries) != 0 {
					t.Fatalf("existing empty destination changed: %v %v", entries, err)
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("failed new project left destination: %v", err)
			}
			siblings, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range siblings {
				if strings.HasPrefix(entry.Name(), ".keel-new-") {
					t.Fatal("staging directory leaked")
				}
			}
		})
	}
}

func TestNewDestinationReadErrorReturnsImmediately(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "not-a-directory")
	sentinel := []byte("user data")
	if err := os.WriteFile(target, sentinel, 0644); err != nil {
		t.Fatal(err)
	}
	_, want := os.ReadDir(target)
	if want == nil || os.IsNotExist(want) {
		t.Fatal("fixture must produce non-notexist ReadDir error")
	}
	c, _ := newCLI(root)
	// Invalid checkout would fail later; the ReadDir error must win first.
	err := c.newProject([]string{target, "-replace", filepath.Join(root, "missing-checkout")})
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || pathErr.Path != target {
		t.Fatalf("expected initial ReadDir error, got %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != string(sentinel) {
		t.Fatal("user file changed", err)
	}
}

func TestInstallProjectPreservesConcurrentFiles(t *testing.T) {
	root := t.TempDir()
	stage := filepath.Join(root, "stage")
	target := filepath.Join(root, "target")
	os.Mkdir(stage, 0755)
	os.Mkdir(target, 0755)
	os.WriteFile(filepath.Join(stage, "main.go"), []byte("generated"), 0644)
	user := filepath.Join(target, "main.go")
	os.WriteFile(user, []byte("concurrent user data"), 0644)
	if err := installProject(stage, target); err == nil {
		t.Fatal("nonempty concurrent destination accepted")
	}
	got, err := os.ReadFile(user)
	if err != nil || string(got) != "concurrent user data" {
		t.Fatal("concurrent data changed", err)
	}
}
