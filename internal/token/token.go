// Package token tokenizes a REPL input line into pipeline segments.
//
// It honours single and double quotes (a backslash inside quotes escapes the
// quote character), splits on unquoted pipe characters, and records which tokens
// carried quotes so later expansion can skip them. It holds no knowledge of the
// REPL or of command semantics; it is the reusable lexer behind the line domain.
package token

import (
	"strings"

	"github.com/yupsh/repl/internal/constants"
)

// Line is one line of input typed at the prompt.
type Line string

// Token is one tokenized word plus whether any of it was quoted. Quoting
// suppresses tilde and glob expansion, exactly as in a POSIX shell.
type Token struct {
	Text     string
	IsQuoted bool
}

// Segment is the unexpanded token list of one pipeline stage.
type Segment []Token

// Parse tokenizes a command line into pipeline segments. Empty segments are
// preserved so the caller can report an explicit "empty command" for inputs like
// "a |".
func Parse(line Line) ([]Segment, error) {
	t := &tokenizer{}
	for _, ch := range string(line) {
		t.feed(ch)
	}
	return t.finish()
}

// tokenizer accumulates tokens, the current segment, and quote state across a
// single line. Pointer receiver: it is a single-pass mutable accumulator.
type tokenizer struct {
	token     strings.Builder
	segments  []Segment
	current   Segment
	quote     rune
	hasToken  bool
	isQuoted  bool
	isInQuote bool
	isEscaped bool
}

// feed consumes one rune, updating tokenizer state.
func (t *tokenizer) feed(ch rune) {
	switch {
	case t.isEscaped:
		_, _ = t.token.WriteRune(ch)
		t.hasToken = true
		t.isEscaped = false
	case t.isInQuote:
		t.inQuoteFeed(ch)
	default:
		t.bareFeed(ch)
	}
}

// inQuoteFeed handles a rune while inside a quoted span.
func (t *tokenizer) inQuoteFeed(ch rune) {
	switch ch {
	case '\\':
		t.isEscaped = true
	case t.quote:
		t.isInQuote = false
	default:
		_, _ = t.token.WriteRune(ch)
		t.hasToken = true
	}
}

// bareFeed handles a rune outside any quoted span.
func (t *tokenizer) bareFeed(ch rune) {
	switch ch {
	case '\'', '"':
		t.isInQuote = true
		t.quote = ch
		t.hasToken = true
		t.isQuoted = true
	case ' ', '\t':
		t.flushToken()
	case '|':
		t.flushSegment()
	default:
		_, _ = t.token.WriteRune(ch)
		t.hasToken = true
	}
}

// flushToken appends the pending token to the current segment, if any.
func (t *tokenizer) flushToken() {
	if !t.hasToken {
		return
	}
	t.current = append(t.current, Token{Text: t.token.String(), IsQuoted: t.isQuoted})
	t.token.Reset()
	t.hasToken = false
	t.isQuoted = false
}

// flushSegment closes the current segment at a pipe boundary.
func (t *tokenizer) flushSegment() {
	t.flushToken()
	t.segments = append(t.segments, t.current)
	t.current = nil
}

// finish flushes the trailing token and segment and rejects an open quote.
func (t *tokenizer) finish() ([]Segment, error) {
	if t.isInQuote || t.isEscaped {
		return nil, constants.ErrUnterminatedQuote
	}
	t.flushSegment()
	return t.segments, nil
}
