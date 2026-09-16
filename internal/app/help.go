package app

import (
	"sort"

	"github.com/yupsh/repl/internal/command"
)

// help prints the built-in list, the command catalogue, and usage examples.
func (e Session) help() {
	w := e.out
	reg := command.Registry()
	fprintln(w, "yupsh REPL — available commands")
	fprintln(w)
	fprintln(w, "Built-ins:")
	fprintln(w, "  help              show this message")
	fprintln(w, "  version           show version information")
	fprintln(w, "  clear             clear the screen")
	fprintln(w, "  exit, quit        leave the REPL")
	fprintln(w)
	fprintln(w, "Commands:")
	for _, name := range sortedNames(reg) {
		fprintf(w, "  %-9s %s\n", name, reg[command.Name(name)].Summary)
	}
	fprintln(w)
	fprintln(w, "Pipelines: command1 | command2 | command3")
	fprintln(w, "Examples:")
	fprintln(w, "  echo hello world")
	fprintln(w, "  seq 1 10 | grep 5")
	fprintln(w, "  seq 1 100 | wc -l")
	fprintln(w)
}

// sortedNames returns the registry's command names in lexical order.
func sortedNames(reg map[command.Name]command.Builder) []string {
	names := make([]string, 0, len(reg))
	for name := range reg {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return names
}
