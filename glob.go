package repl

import (
	"strings"

	"github.com/spf13/afero"
)

// expandArgs applies shell-style expansion to a command's argument tokens:
// tilde expansion then pathname globbing, on unquoted tokens only. A glob with
// no matches is left literal (POSIX default, no "nullglob"), so non-matching
// patterns reach the command unchanged.
func expandArgs(env environment, tokens []rawToken) Argv {
	var out Argv
	for _, tok := range tokens {
		for _, word := range expandToken(env, tok) {
			out = append(out, Argument(word))
		}
	}
	return out
}

// expandToken expands a single token into zero or more words.
func expandToken(env environment, tok rawToken) []string {
	if tok.quoted {
		return []string{tok.text}
	}
	word := tildeExpand(tok.text, env.home)
	if !hasGlobMeta(word) {
		return []string{word}
	}
	matches, err := afero.Glob(env.fs, word)
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
