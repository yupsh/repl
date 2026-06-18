package repl

import (
	"testing"

	"github.com/spf13/afero"
)

func TestTildeExpand(t *testing.T) {
	cases := []struct {
		word, home, want string
	}{
		{"~", "/home/u", "/home/u"},
		{"~/docs", "/home/u", "/home/u/docs"},
		{"~user", "/home/u", "~user"}, // only bare ~ and ~/ expand
		{"plain", "/home/u", "plain"},
		{"~", "", "~"}, // no home injected: left literal
	}
	for _, c := range cases {
		if got := tildeExpand(c.word, c.home); got != c.want {
			t.Errorf("tildeExpand(%q, %q) = %q, want %q", c.word, c.home, got, c.want)
		}
	}
}

func TestHasGlobMeta(t *testing.T) {
	for _, w := range []string{"*.go", "a?b", "x[0-9]"} {
		if !hasGlobMeta(w) {
			t.Errorf("hasGlobMeta(%q) = false", w)
		}
	}
	for _, w := range []string{"plain", "a.go", "~/x"} {
		if hasGlobMeta(w) {
			t.Errorf("hasGlobMeta(%q) = true", w)
		}
	}
}

// globFs builds an in-memory tree used by the expansion tests.
func globFs(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	for path, body := range map[string]string{
		"/p/a.go":      "1\n2\n",
		"/p/b.go":      "3\n",
		"/p/readme.md": "x\n",
	} {
		if err := afero.WriteFile(fs, path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return fs
}

func TestGlobExpansion(t *testing.T) {
	fs := globFs(t)
	cases := []struct {
		line string
		want string
	}{
		{"wc -l /p/*.go", "3\n"},             // glob feeds files to a filter
		{"ls /p/*.go", "/p/a.go\n/p/b.go\n"}, // glob → names
		{"ls '/p/*.go'", "/p/*.go\n"},        // quoted → literal, no expansion
		{"echo /p/*.xml", "/p/*.xml\n"},      // no match → literal
		{"echo /p/[bad", "/p/[bad\n"},        // malformed pattern → literal
		{"ls /p", "a.go\nb.go\nreadme.md\n"}, // single directory → list entries
		{"ls /p/a.go", "/p/a.go\n"},          // single non-directory → echo name
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("execute(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestTildeInPipeline(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, testHome+"/data.txt", []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := mustExec(t, fs, "wc -l ~/data.txt"); got != "3\n" {
		t.Errorf("wc -l ~/data.txt = %q, want %q", got, "3\n")
	}
}
