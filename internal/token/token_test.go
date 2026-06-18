package token

import (
	"errors"
	"testing"

	"github.com/yupsh/repl/internal/constants"
)

func TestParse(t *testing.T) {
	cases := []struct {
		line string
		want [][]string
	}{
		{"a b c", [][]string{{"a", "b", "c"}}},
		{"echo \"hello world\"", [][]string{{"echo", "hello world"}}},
		{"echo 'a b'", [][]string{{"echo", "a b"}}},
		{`echo "a\"b"`, [][]string{{"echo", `a"b`}}},
		{"a | b | c", [][]string{{"a"}, {"b"}, {"c"}}},
		{"a |", [][]string{{"a"}, {}}},
		{"  spaced   out  ", [][]string{{"spaced", "out"}}},
		{`echo ""`, [][]string{{"echo", ""}}},
	}
	for _, c := range cases {
		segs, err := Parse(Line(c.line))
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.line, err)
		}
		if !segmentsEqual(segs, c.want) {
			t.Errorf("Parse(%q) = %v, want %v", c.line, segs, c.want)
		}
	}
}

func TestParseUnterminated(t *testing.T) {
	for _, line := range []string{`echo "open`, `echo "trailing\`} {
		if _, err := Parse(Line(line)); !errors.Is(err, constants.ErrUnterminatedQuote) {
			t.Errorf("Parse(%q) err = %v, want ErrUnterminatedQuote", line, err)
		}
	}
}

// segmentsEqual compares parsed segments against expected string slices.
func segmentsEqual(segs []Segment, want [][]string) bool {
	if len(segs) != len(want) {
		return false
	}
	for i, seg := range segs {
		if len(seg) != len(want[i]) {
			return false
		}
		for j, tok := range seg {
			if tok.Text != want[i][j] {
				return false
			}
		}
	}
	return true
}
