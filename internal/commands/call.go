package commands

import (
	"encoding/json"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CallCmd invokes an RPC procedure on the mesh.
type CallCmd struct{}

func (c *CallCmd) Name() string        { return "call" }
func (c *CallCmd) Aliases() []string   { return []string{"rpc"} }
func (c *CallCmd) Description() string { return "Call a mesh procedure (/call <mri> [json-args])" }

func (c *CallCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			s := ctx.Styles
			return InjectSystemMsg{
				Content: s.Error.Render("Usage: /call <procedure-mri> [json-args]") + "\n" +
					s.Subtle.Render("Example: /call mri:proc:io.macula/echo {\"msg\":\"hello\"}"),
			}
		}
	}

	procedure := args[0]
	var rpcArgs interface{}

	// Parse optional JSON args
	if len(args) > 1 {
		jsonStr := strings.Join(args[1:], " ")
		if err := json.Unmarshal([]byte(jsonStr), &rpcArgs); err != nil {
			return func() tea.Msg {
				return InjectSystemMsg{
					Content: ctx.Styles.Error.Render("Invalid JSON args: " + err.Error()),
				}
			}
		}
	}

	return func() tea.Msg {
		s := ctx.Styles

		result, err := ctx.Client.RPCCall(procedure, rpcArgs)
		if err != nil {
			return InjectSystemMsg{
				Content: s.Error.Render("RPC Error: " + err.Error()),
			}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("RPC Result"))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("  Procedure: "))
		b.WriteString(s.CardValue.Render(procedure))
		b.WriteString("\n")

		if result.Duration != "" {
			b.WriteString(s.CardLabel.Render("  Duration:  "))
			b.WriteString(s.Subtle.Render(result.Duration))
			b.WriteString("\n")
		}

		if result.Error != "" {
			b.WriteString(s.CardLabel.Render("  Error:     "))
			b.WriteString(s.Error.Render(result.Error))
		} else {
			b.WriteString("\n")
			b.WriteString(s.Bold.Render("  Result:"))
			b.WriteString("\n")

			// Pretty-print JSON result
			var pretty json.RawMessage
			if json.Unmarshal(result.Result, &pretty) == nil {
				formatted, err := json.MarshalIndent(pretty, "  ", "  ")
				if err == nil {
					b.WriteString("  ")
					b.WriteString(s.CardValue.Render(string(formatted)))
				} else {
					b.WriteString("  ")
					b.WriteString(s.CardValue.Render(string(result.Result)))
				}
			} else {
				b.WriteString("  ")
				b.WriteString(s.CardValue.Render(string(result.Result)))
			}
		}

		return InjectSystemMsg{Content: b.String()}
	}
}
