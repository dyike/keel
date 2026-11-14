package kit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffPatchAppliesWithGit(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	cases := [][2]string{{"", "new\n"}, {"old\n", ""}, {"old", "new"}, {"old\r\nkeep\r\n", "new\r\nkeep\r\n"}, {"a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\n", "A\nb\nc\nd\ne\nf\ng\nh\ni\nj\nK\nl\n"}}
	for _, tc := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "file"), []byte(tc[0]), 0600); err != nil {
			t.Fatal(err)
		}
		patch := Diff(tc[0], tc[1]).Patch("file", "file")
		cmd := exec.Command(git, "-c", "core.autocrlf=false", "-c", "apply.whitespace=nowarn", "apply", "--no-index", "-")
		cmd.Dir = dir
		cmd.Stdin = strings.NewReader(patch)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git apply: %v\n%s\n%s", err, out, patch)
		}
		got, err := os.ReadFile(filepath.Join(dir, "file"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != tc[1] {
			t.Fatalf("patch result %q, want %q", got, tc[1])
		}
	}
}
