package app

import (
	"fmt"
	"sort"

	"github.com/yupsh/repl/internal/command"
)

// help prints the built-in list, the command catalogue, and usage examples.
func (e Session) help() {
	w := e.out
	reg := command.Registry()
	fmt.Fprintln(w, "yupsh REPL — available commands")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Built-ins:")
	fmt.Fprintln(w, "  help              show this message")
	fmt.Fprintln(w, "  version           show version information")
	fmt.Fprintln(w, "  clear             clear the screen")
	fmt.Fprintln(w, "  exit, quit        leave the REPL")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, name := range sortedNames(reg) {
		fmt.Fprintf(w, "  %-9s %s\n", name, reg[command.Name(name)].Summary)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Pipelines: command1 | command2 | command3")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  echo hello world")
	fmt.Fprintln(w, "  seq 1 10 | grep 5")
	fmt.Fprintln(w, "  seq 1 100 | wc -l")
	fmt.Fprintln(w)
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
