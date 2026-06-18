// Package pipeline assembles built command segments into a runnable typed-stream
// pipeline and executes it against the gloo-foo/framework.
//
// Plan folds an ordered list of stages into an Assembly: the first stage chooses
// the input source (a source command, named files, or stdin) and later stages
// must be plain filters. Assembly.Run streams the result to an output writer.
// The package knows how to wire and run segments; it does not know how a segment
// is built (that is internal/command) nor how a line is parsed (the line domain).
package pipeline

import (
	"context"
	"io"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/command"
	"github.com/yupsh/repl/internal/constants"
)

// Stage pairs a built segment with the command name that produced it, so
// assembly errors can name the offending command.
type Stage struct {
	Name    command.Name
	Segment command.Segment
}

// Assembly is the resolved plan for a pipeline: an input source and the ordered
// transform commands applied to it. The add* methods take a pointer receiver
// because they accumulate the source and command list as stages are folded in;
// Run is a value receiver as it only reads the finished plan.
type Assembly struct {
	source gloo.Source[[]byte]
	cmds   []any
}

// Plan turns built stages into an Assembly. The first stage chooses the input
// source (a source command, named files, or stdin); later stages must be plain
// filters.
func Plan(stages []Stage, fs afero.Fs, stdin io.Reader) (Assembly, error) {
	a := Assembly{source: gloo.ByteReaderSource([]io.Reader{stdin})}
	for i, s := range stages {
		if err := a.add(i, s, fs); err != nil {
			return Assembly{}, err
		}
	}
	return a, nil
}

// add folds one stage into the assembly.
func (a *Assembly) add(index int, s Stage, fs afero.Fs) error {
	if index == 0 {
		a.addFirst(s, fs)
		return nil
	}
	return a.addNext(s)
}

// addFirst resolves the input-selecting first stage.
func (a *Assembly) addFirst(s Stage, fs afero.Fs) {
	switch {
	case s.Segment.Source != nil:
		a.source = s.Segment.Source
	case s.Segment.Inputs != nil:
		a.source = gloo.SliceSource(s.Segment.Inputs)
		a.cmds = append(a.cmds, s.Segment.Command)
	case len(s.Segment.Files) > 0:
		a.source = gloo.ByteFileSource(fs, s.Segment.Files)
		a.cmds = append(a.cmds, s.Segment.Command)
	default:
		a.cmds = append(a.cmds, s.Segment.Command)
	}
}

// addNext folds a non-first stage, rejecting sources and positional arguments
// (which would need to source input the pipeline already provides).
func (a *Assembly) addNext(s Stage) error {
	if s.Segment.Source != nil {
		return constants.ErrSourceMidPipeline.With(nil, string(s.Name))
	}
	if len(s.Segment.Files) > 0 || s.Segment.Inputs != nil {
		return constants.ErrArgsMidPipeline.With(nil, string(s.Name))
	}
	a.cmds = append(a.cmds, s.Segment.Command)
	return nil
}

// Run executes the assembled pipeline, writing output to out.
func (a Assembly) Run(ctx context.Context, out io.Writer) error {
	_, err := gloo.RunContext(ctx, a.source, gloo.ByteWriteTo(out), a.cmds...)
	return err
}
