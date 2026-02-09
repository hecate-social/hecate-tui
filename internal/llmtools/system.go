package llmtools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// UserQuestionHandler is called when ask_user tool is invoked.
// The TUI must set this to handle user prompts.
type UserQuestionHandler func(question string, options []string) (string, error)

var userQuestionHandler UserQuestionHandler

// SetUserQuestionHandler sets the callback for ask_user tool.
func SetUserQuestionHandler(h UserQuestionHandler) {
	userQuestionHandler = h
}

// RegisterSystemTools adds system interaction tools to the registry.
func RegisterSystemTools(r *Registry) {
	r.Register(runCommandTool(), runCommandHandler)
	r.Register(askUserTool(), askUserHandler)
	r.Register(getEnvTool(), getEnvHandler)
	r.Register(cwdTool(), cwdHandler)
}

// --- run_command ---

func runCommandTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("command", String("The shell command to execute"))
	params.AddProperty("working_dir", String("Working directory for the command (default: current directory)"))
	params.AddProperty("timeout", Integer("Timeout in seconds (default: 60, max: 300)"))
	params.AddRequired("command")

	return Tool{
		Name:             "run_command",
		Description:      "Execute a shell command. For safety, certain destructive commands are blocked.",
		Parameters:       params,
		Category:         CategorySystem,
		RequiresApproval: true,
	}
}

type runCommandArgs struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
	Timeout    int    `json:"timeout"`
}

func runCommandHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a runCommandArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	// Default timeout
	timeout := a.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	if timeout > 300 {
		timeout = 300
	}

	workingDir := a.WorkingDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "/"
		}
	}
	workingDir = expandHomePath(workingDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute via shell
	cmd := exec.CommandContext(ctx, "sh", "-c", a.Command)
	cmd.Dir = workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("$ %s\n", a.Command))
	sb.WriteString(fmt.Sprintf("Working directory: %s\n", workingDir))
	sb.WriteString(fmt.Sprintf("Duration: %.2fs\n\n", duration.Seconds()))

	if stdout.Len() > 0 {
		output := stdout.String()
		// Truncate if too long
		if len(output) > 10000 {
			output = output[:10000] + "\n... (truncated, " + fmt.Sprintf("%d", len(stdout.String())-10000) + " bytes omitted)"
		}
		sb.WriteString("STDOUT:\n")
		sb.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			sb.WriteString("\n")
		}
	}

	if stderr.Len() > 0 {
		output := stderr.String()
		if len(output) > 5000 {
			output = output[:5000] + "\n... (truncated)"
		}
		sb.WriteString("\nSTDERR:\n")
		sb.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			sb.WriteString("\n")
		}
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sb.WriteString(fmt.Sprintf("\nCommand timed out after %d seconds", timeout))
		} else {
			sb.WriteString(fmt.Sprintf("\nExit error: %s", err.Error()))
		}
	} else {
		sb.WriteString("\nExit code: 0")
	}

	return sb.String(), nil
}

// --- ask_user ---

func askUserTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("question", String("The question to ask the user"))
	params.AddProperty("options", ParameterSpec{
		Type:        "array",
		Description: "Optional list of choices to present (if not provided, user can type freely)",
	})
	params.AddRequired("question")

	return Tool{
		Name:             "ask_user",
		Description:      "Ask the user a question and wait for their response. Use when you need clarification or input.",
		Parameters:       params,
		Category:         CategorySystem,
		RequiresApproval: false,
	}
}

type askUserArgs struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func askUserHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a askUserArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Question == "" {
		return "", fmt.Errorf("question is required")
	}

	if userQuestionHandler == nil {
		return "", fmt.Errorf("ask_user is not configured (no handler set)")
	}

	answer, err := userQuestionHandler(a.Question, a.Options)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User answered: %s", answer), nil
}

// --- get_env ---

func getEnvTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("name", String("Environment variable name"))
	params.AddRequired("name")

	return Tool{
		Name:             "get_env",
		Description:      "Get the value of an environment variable.",
		Parameters:       params,
		Category:         CategorySystem,
		RequiresApproval: false,
	}
}

type getEnvArgs struct {
	Name string `json:"name"`
}

func getEnvHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a getEnvArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	// Block sensitive environment variables
	sensitive := map[string]bool{
		"AWS_SECRET_ACCESS_KEY": true,
		"AWS_SESSION_TOKEN":     true,
		"GITHUB_TOKEN":          true,
		"GH_TOKEN":              true,
		"NPM_TOKEN":             true,
		"OPENAI_API_KEY":        true,
		"ANTHROPIC_API_KEY":     true,
		"DATABASE_URL":          true,
		"DB_PASSWORD":           true,
	}

	upperName := strings.ToUpper(a.Name)
	if sensitive[upperName] || strings.Contains(upperName, "SECRET") || strings.Contains(upperName, "PASSWORD") || strings.Contains(upperName, "TOKEN") || strings.Contains(upperName, "API_KEY") {
		return "", fmt.Errorf("access to sensitive environment variable '%s' is blocked", a.Name)
	}

	value := os.Getenv(a.Name)
	if value == "" {
		return fmt.Sprintf("Environment variable '%s' is not set", a.Name), nil
	}

	return fmt.Sprintf("%s=%s", a.Name, value), nil
}

// --- cwd ---

func cwdTool() Tool {
	params := NewObjectParameters()

	return Tool{
		Name:             "cwd",
		Description:      "Get the current working directory.",
		Parameters:       params,
		Category:         CategorySystem,
		RequiresApproval: false,
	}
}

func cwdHandler(ctx context.Context, args json.RawMessage) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return cwd, nil
}
