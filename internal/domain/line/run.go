package line

import (
	"io"

	"github.com/yupsh/repl/internal/command"
	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/expansion"
	"github.com/yupsh/repl/internal/flags"
	"github.com/yupsh/repl/internal/pipeline"
	"github.com/yupsh/repl/internal/token"
)

// Run resolves the input line into a planned pipeline: it tokenizes the line,
// builds each segment against the command registry, and folds the stages into a
// pipeline.Assembly the caller streams to its output. It holds no presentation
// or stream-running logic.
func Run(cfg Config, stdin io.Reader, in token.Line) (pipeline.Assembly, error) {
	segments, err := token.Parse(in)
	if err != nil {
		return pipeline.Assembly{}, err
	}
	stages, err := buildStages(cfg, command.Registry(), segments)
	if err != nil {
		return pipeline.Assembly{}, err
	}
	return pipeline.Plan(stages, cfg.Fs, stdin)
}

// buildStages builds every pipeline segment.
func buildStages(cfg Config, reg map[command.Name]command.Builder, segments []token.Segment) ([]pipeline.Stage, error) {
	stages := make([]pipeline.Stage, 0, len(segments))
	for _, s := range segments {
		st, err := buildStage(cfg, reg, s)
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
}

// buildStage resolves and builds a single segment.
func buildStage(cfg Config, reg map[command.Name]command.Builder, s token.Segment) (pipeline.Stage, error) {
	if len(s) == 0 {
		return pipeline.Stage{}, constants.ErrEmptyCommand
	}
	name := command.Name(s[0].Text)
	b, ok := reg[name]
	if !ok {
		return pipeline.Stage{}, constants.ErrUnknownCommand.With(nil, string(name))
	}
	opts, positional, err := resolveArgs(b, expand(cfg, s[1:]))
	if err != nil {
		return pipeline.Stage{}, err
	}
	seg, err := b.Build(cfg.Fs, positional, opts)
	if err != nil {
		return pipeline.Stage{}, err
	}
	return pipeline.Stage{Name: name, Segment: seg}, nil
}

// expand applies tilde and glob expansion to a segment's argument tokens.
func expand(cfg Config, tokens []token.Token) flags.Argv {
	words := expansion.Expand(cfg.Fs, cfg.Home, tokens)
	argv := make(flags.Argv, len(words))
	for i, w := range words {
		argv[i] = flags.Argument(w)
	}
	return argv
}

// resolveArgs splits a segment's arguments into options and positionals,
// honouring a builder's raw passthrough mode.
func resolveArgs(b command.Builder, args flags.Argv) (opts []any, positional flags.Argv, err error) {
	if b.Raw {
		return nil, args, nil
	}
	return flags.Parse(b.Flags, args)
}
