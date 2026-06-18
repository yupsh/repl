package expansion

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/token"
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
		"/p/a.go":      "1\n",
		"/p/b.go":      "2\n",
		"/p/readme.md": "x\n",
	} {
		if err := afero.WriteFile(fs, path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return fs
}

func TestExpand(t *testing.T) {
	fs := globFs(t)
	const home Home = "/home/u"
	cases := []struct {
		name   string
		tokens []token.Token
		want   []string
	}{
		{"quoted literal", []token.Token{{Text: "/p/*.go", Quoted: true}}, []string{"/p/*.go"}},
		{"tilde", []token.Token{{Text: "~"}}, []string{"/home/u"}},
		{"plain word", []token.Token{{Text: "plain"}}, []string{"plain"}},
		{"glob match", []token.Token{{Text: "/p/*.go"}}, []string{"/p/a.go", "/p/b.go"}},
		{"glob no match", []token.Token{{Text: "/p/*.xml"}}, []string{"/p/*.xml"}},
		{"glob malformed", []token.Token{{Text: "/p/[bad"}}, []string{"/p/[bad"}},
		{"multiple tokens", []token.Token{{Text: "~"}, {Text: "plain"}}, []string{"/home/u", "plain"}},
	}
	for _, c := range cases {
		got := Expand(fs, home, c.tokens)
		if !equal(got, c.want) {
			t.Errorf("%s: Expand = %v, want %v", c.name, got, c.want)
		}
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
