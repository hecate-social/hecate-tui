package commands

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Registry holds all registered commands and handles dispatch.
type Registry struct {
	commands map[string]Command   // name → command
	aliases  map[string]string    // alias → canonical name
	ordered  []string             // sorted command names for display
}

// NewRegistry creates a registry with all built-in commands registered.
func NewRegistry() *Registry {
	r := &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}

	// Register built-in commands
	r.Register(&HelpCmd{registry: r})
	r.Register(&HistoryCmd{})
	r.Register(&ClearCmd{})
	r.Register(&DeleteCmd{})
	r.Register(&QuitCmd{})
	r.Register(&StatusCmd{})
	r.Register(&HealthCmd{})
	r.Register(&GeoCmd{})
	r.Register(&ModelsCmd{})
	r.Register(&ModelCmd{})
	r.Register(&LoadCmd{})
	r.Register(&MeCmd{})
	r.Register(&NewCmd{})
	r.Register(&BrowseCmd{})
	r.Register(&CallCmd{})
	r.Register(&ConfigCmd{})
	r.Register(&EditCmd{})
	r.Register(&FindCmd{})
	r.Register(&PairCmd{})
	r.Register(&ProjectCmd{})
	r.Register(&SaveCmd{})
	r.Register(&SubscriptionsCmd{})
	r.Register(&SystemCmd{})
	r.Register(&ThemeCmd{})
	r.Register(&ToolsCmd{})
	r.Register(&LLMToolsCmd{})
	r.Register(&CartwheelCmd{})
	r.Register(&ProviderCmd{})
	r.Register(&RoleCmd{})
	r.Register(&AboutCmd{})
	r.Register(&TorchCmd{})
	r.Register(&TorchesCmd{})
	r.Register(&ChatCmd{})
	r.Register(&BackCmd{})
	r.Register(&CartwheelsCmd{})
	r.Register(&AgentsCmd{})
	r.Register(&CostCmd{})

	return r
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	name := cmd.Name()
	r.commands[name] = cmd
	for _, alias := range cmd.Aliases() {
		r.aliases[alias] = name
	}
	r.ordered = append(r.ordered, name)
	sort.Strings(r.ordered)
}

// Dispatch parses and executes a command string.
// Returns a tea.Cmd that should be batched into the update loop.
func (r *Registry) Dispatch(input string, ctx *Context) tea.Cmd {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Strip leading / or : prefix
	if input[0] == '/' || input[0] == ':' {
		input = input[1:]
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	name := strings.ToLower(parts[0])
	args := parts[1:]

	// Look up by name first, then alias
	cmd, ok := r.commands[name]
	if !ok {
		canonical, aliasOk := r.aliases[name]
		if aliasOk {
			cmd = r.commands[canonical]
		}
	}

	if cmd == nil {
		return func() tea.Msg {
			return InjectSystemMsg{Content: "Unknown command: " + name + "\nType /help for available commands."}
		}
	}

	return cmd.Execute(args, ctx)
}

// Complete returns command names that match the given prefix.
func (r *Registry) Complete(prefix string) []string {
	prefix = strings.ToLower(strings.TrimLeft(prefix, "/:"))
	if prefix == "" {
		return r.ordered
	}

	var matches []string
	for _, name := range r.ordered {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	// Also check aliases
	for alias, canonical := range r.aliases {
		if strings.HasPrefix(alias, prefix) {
			// Avoid duplicates
			found := false
			for _, m := range matches {
				if m == canonical {
					found = true
					break
				}
			}
			if !found {
				matches = append(matches, alias)
			}
		}
	}

	sort.Strings(matches)
	return matches
}

// CompleteWithArgs returns completions for commands with arguments.
// input is the full command line (e.g., "torch loc" or "t 1").
// ctx is needed for commands that fetch data for completion.
func (r *Registry) CompleteWithArgs(input string, ctx *Context) []string {
	input = strings.TrimLeft(input, "/:")
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return r.ordered
	}

	cmdName := strings.ToLower(parts[0])

	// If only command name typed (no space after), complete the command name
	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		return r.Complete(cmdName)
	}

	// Find the command
	cmd, ok := r.commands[cmdName]
	if !ok {
		// Check aliases
		if canonical, aliasOk := r.aliases[cmdName]; aliasOk {
			cmd = r.commands[canonical]
		}
	}

	if cmd == nil {
		return nil
	}

	// Check if command supports completion
	completable, ok := cmd.(Completable)
	if !ok {
		return nil
	}

	// Get the arguments (everything after command name)
	args := parts[1:]

	// If input ends with space, user is starting a new argument
	if strings.HasSuffix(input, " ") {
		args = append(args, "")
	}

	return completable.Complete(args, ctx)
}

// List returns all commands in sorted order.
func (r *Registry) List() []Command {
	var cmds []Command
	for _, name := range r.ordered {
		cmds = append(cmds, r.commands[name])
	}
	return cmds
}
