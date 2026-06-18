package app

import (
	"bufio"
	"fmt"
	"io"
)

// Prompt is printed before each input line.
const Prompt = "yup> "

// LineReader supplies successive command lines to the REPL, returning io.EOF
// when input is exhausted. It is the seam that lets interactive line editing and
// history be supplied by the terminal: *golang.org/x/term.Terminal already
// satisfies this interface, so the main program wires one in for TTY sessions
// while tests and piped input use the plain scanner reader.
type LineReader interface {
	ReadLine() (string, error)
}

// scanReader is the non-interactive LineReader: it prints the prompt and reads
// newline-delimited lines from a reader, with no terminal editing. Pointer
// receiver: it wraps a stateful bufio.Scanner.
type scanReader struct {
	scanner *bufio.Scanner
	out     io.Writer
}

// newScanReader builds a scanReader over in, printing prompts to out. The scan
// buffer grows up to 1 MiB so long lines are not truncated.
func newScanReader(in io.Reader, out io.Writer) *scanReader {
	s := bufio.NewScanner(in)
	s.Buffer(make([]byte, 0, 64*1024), 1<<20)
	return &scanReader{scanner: s, out: out}
}

// ReadLine prints the prompt and returns the next line, or io.EOF at end.
func (r *scanReader) ReadLine() (string, error) {
	fmt.Fprint(r.out, Prompt)
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}
