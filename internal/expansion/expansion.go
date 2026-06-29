// Package expansion applies shell-style expansion to command argument tokens:
// tilde expansion then pathname globbing, on unquoted tokens only.
//
// A glob with no matches is left literal (POSIX default, no "nullglob"), so
// non-matching patterns reach the command unchanged. Expansion resolves globs
// against an injected afero.Fs, so it stays testable against an in-memory tree;
// it holds no knowledge of the REPL or of command semantics.
package expansion

import (
	"strings"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/token"
)

// Home is the home directory used for "~" expansion. An empty value disables
// tilde expansion.
type Home string

// Expand expands a command's argument tokens into words, applying tilde then
// glob expansion to unquoted tokens only.
func Expand(fs afero.Fs, home Home, tokens []token.Token) []string {
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		out = append(out, expandToken(fs, home, tok)...)
	}
	return out
}

// expandToken expands a single token into zero or more words.
func expandToken(fs afero.Fs, home Home, tok token.Token) []string {
	if tok.Quoted {
		return []string{tok.Text}
	}
	word := tildeExpand(tok.Text, string(home))
	if !hasGlobMeta(word) {
		return []string{word}
	}
	matches, err := afero.Glob(fs, word)
	if err != nil || len(matches) == 0 {
		return []string{word}
	}
	return matches
}

// tildeExpand replaces a leading "~" or "~/" with the home directory. Other
// forms (including "~user") are left untouched.
func tildeExpand(word, home string) string {
	if home == "" {
		return word
	}
	if word == "~" {
		return home
	}
	if strings.HasPrefix(word, "~/") {
		return home + word[1:]
	}
	return word
}

// hasGlobMeta reports whether a word contains pathname-glob metacharacters.
func hasGlobMeta(word string) bool {
	return strings.ContainsAny(word, "*?[")
}
