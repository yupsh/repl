package command

import (
	"context"

	xargs "github.com/gloo-foo/cmd-xargs"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/flags"
)

// buildXargs builds the xargs segment. With no positional command it is a plain
// regrouping filter (-n groups fields into argument lines). With a positional
// command it runs that command once per argument group, resolving each group's
// argv against the registry (and falling back to a subprocess for unknown
// commands) via an injected XargsExec factory.
func buildXargs(fs afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	if len(positional) == 0 {
		return Segment{Command: xargs.Xargs(opts...)}, nil
	}
	args := make([]any, 0, len(positional)+len(opts)+1)
	for _, p := range positional {
		args = append(args, gloo.File(p))
	}
	args = append(args, opts...)
	args = append(args, xargs.XargsExec(xargsFactory(fs)))
	return Segment{Command: xargs.Xargs(args...)}, nil
}

// xargsFactory builds the per-group command factory: registry first, then a
// subprocess for any command the registry does not know.
func xargsFactory(fs afero.Fs) xargs.CommandFor {
	reg := Registry()
	return func(argv []string) gloo.Command[[]byte, []byte] {
		cmd, err := commandFromArgv(fs, reg, argv)
		if err != nil {
			return xargs.Subprocess(argv)
		}
		return cmd
	}
}

// commandFromArgv builds a runnable command from a group's argv by resolving it
// against the registry exactly as a typed pipeline stage: argv[0] is the command
// keyword and argv[1:] its arguments.
func commandFromArgv(fs afero.Fs, reg map[Name]Builder, argv []string) (gloo.Command[[]byte, []byte], error) {
	if len(argv) == 0 {
		return nil, constants.ErrEmptyCommand
	}
	b, ok := reg[Name(argv[0])]
	if !ok {
		return nil, constants.ErrUnknownCommand.With(nil, argv[0])
	}
	opts, positional, err := resolveArgv(b, toArgv(argv[1:]))
	if err != nil {
		return nil, err
	}
	seg, err := b.Build(fs, positional, opts)
	if err != nil {
		return nil, err
	}
	return segmentCommand(fs, seg), nil
}

// resolveArgv splits an argv tail into options and positionals, honouring a
// builder's raw passthrough mode (mirrors the line domain's argument handling).
func resolveArgv(b Builder, args flags.Argv) (opts []any, positional flags.Argv, err error) {
	if b.Raw {
		return nil, args, nil
	}
	return flags.Parse(b.Flags, args)
}

// toArgv converts a string slice into the flag package's argument vector.
func toArgv(args []string) flags.Argv {
	out := make(flags.Argv, len(args))
	for i, a := range args {
		out[i] = flags.Argument(a)
	}
	return out
}

// segmentCommand adapts a built Segment into a command that ignores its input
// stream and emits the segment's own output — a source, optionally piped through
// the segment's transform. This is the single-stage analogue of pipeline.Plan,
// used to run an xargs child against empty input.
func segmentCommand(fs afero.Fs, seg Segment) gloo.Command[[]byte, []byte] {
	src := segmentSource(fs, seg)
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		stream := src.Stream(ctx)
		if seg.Command != nil {
			stream = seg.Command.Execute(ctx, stream)
		}
		return stream
	})
}

// segmentSource resolves the input a segment sources from: its own source, its
// literal inputs, its named files, or an empty stream.
func segmentSource(fs afero.Fs, seg Segment) gloo.Source[[]byte] {
	switch {
	case seg.Source != nil:
		return seg.Source
	case seg.Inputs != nil:
		return gloo.SliceSource(seg.Inputs)
	case len(seg.Files) > 0:
		return gloo.ByteFileSource(fs, seg.Files)
	default:
		return gloo.SliceSource[[]byte](nil)
	}
}
