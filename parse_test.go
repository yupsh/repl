package repl

import "testing"

func TestParseLine(t *testing.T) {
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
		segs, err := parseLine(Line(c.line))
		if err != nil {
			t.Fatalf("parseLine(%q): %v", c.line, err)
		}
		if !segmentsEqual(segs, c.want) {
			t.Errorf("parseLine(%q) = %v, want %v", c.line, segs, c.want)
		}
	}
}

func TestParseLineUnterminated(t *testing.T) {
	for _, line := range []string{`echo "open`, `echo "trailing\`} {
		if _, err := parseLine(Line(line)); err != ErrUnterminatedQuote {
			t.Errorf("parseLine(%q) err = %v, want ErrUnterminatedQuote", line, err)
		}
	}
}

// segmentsEqual compares built segments against expected string slices.
func segmentsEqual(segs []rawSegment, want [][]string) bool {
	if len(segs) != len(want) {
		return false
	}
	for i, seg := range segs {
		if len(seg) != len(want[i]) {
			return false
		}
		for j, tok := range seg {
			if tok.text != want[i][j] {
				return false
			}
		}
	}
	return true
}
