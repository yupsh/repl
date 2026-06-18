package repl

import (
	"context"
	"io"

	gloo "github.com/gloo-foo/framework"
)

// stage pairs a built segment with the command name that produced it, so
// assembly errors can name the offending command.
type stage struct {
	name    CommandName
	segment segment
}

// assembly is the resolved plan for a pipeline: an input source and the ordered
// transform commands applied to it. The add* methods take a pointer receiver
// because they accumulate the source and command list as stages are folded in;
// run is a value receiver as it only reads the finished plan.
type assembly struct {
	source gloo.Source[[]byte]
	cmds   []any
}

// plan turns built stages into an assembly. The first stage chooses the input
// source (a source command, named files, or stdin); later stages must be plain
// filters.
func plan(stages []stage, env environment, stdin io.Reader) (assembly, error) {
	a := assembly{source: gloo.ByteReaderSource([]io.Reader{stdin})}
	for i, s := range stages {
		if err := a.add(i, s, env); err != nil {
			return assembly{}, err
		}
	}
	return a, nil
}

// add folds one stage into the assembly.
func (a *assembly) add(index int, s stage, env environment) error {
	if index == 0 {
		a.addFirst(s, env)
		return nil
	}
	return a.addNext(s)
}

// addFirst resolves the input-selecting first stage.
func (a *assembly) addFirst(s stage, env environment) {
	switch {
	case s.segment.source != nil:
		a.source = s.segment.source
	case s.segment.inputs != nil:
		a.source = gloo.SliceSource(s.segment.inputs)
		a.cmds = append(a.cmds, s.segment.command)
	case len(s.segment.files) > 0:
		a.source = gloo.ByteFileSource(env.fs, s.segment.files)
		a.cmds = append(a.cmds, s.segment.command)
	default:
		a.cmds = append(a.cmds, s.segment.command)
	}
}

// addNext folds a non-first stage, rejecting sources and positional arguments
// (which would need to source input the pipeline already provides).
func (a *assembly) addNext(s stage) error {
	if s.segment.source != nil {
		return ErrSourceMidPipeline.With(nil, string(s.name))
	}
	if len(s.segment.files) > 0 || s.segment.inputs != nil {
		return ErrArgsMidPipeline.With(nil, string(s.name))
	}
	a.cmds = append(a.cmds, s.segment.command)
	return nil
}

// run executes the assembled pipeline, writing output to out.
func (a assembly) run(ctx context.Context, out io.Writer) error {
	_, err := gloo.RunContext(ctx, a.source, gloo.ByteWriteTo(out), a.cmds...)
	return err
}
