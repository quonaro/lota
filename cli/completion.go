package cli

import (
	"fmt"
	"lota/config"
	"os"
	"path/filepath"
	"strconv"

	"github.com/posener/complete/v2"
	"github.com/posener/complete/v2/predict"
)

// RunCompletion runs shell completion based on COMP_LINE and COMP_POINT env vars.
func RunCompletion() {
	// posener/complete requires COMP_POINT; default to end of line if missing
	if os.Getenv("COMP_POINT") == "" {
		if line := os.Getenv("COMP_LINE"); line != "" {
			_ = os.Setenv("COMP_POINT", strconv.Itoa(len(line)))
		}
	}

	cfg, err := LoadConfig("")
	if err != nil {
		cfg = &config.AppConfig{}
	}

	comp := BuildCompletion(cfg)
	comp.Complete("lota")
}

// BuildCompletion creates a completion tree from the app configuration.
func BuildCompletion(cfg *config.AppConfig) *complete.Command {
	cmd := &complete.Command{
		Sub:   make(map[string]*complete.Command),
		Flags: make(map[string]complete.Predictor),
	}

	// Global flags
	addGlobalFlags(cmd)

	// Top-level groups and commands
	for i := range cfg.Groups {
		g := &cfg.Groups[i]
		cmd.Sub[g.Name] = buildGroupCompletion(g)
	}
	for i := range cfg.Commands {
		c := &cfg.Commands[i]
		cmd.Sub[c.Name] = buildCommandCompletion(c)
	}

	return cmd
}

func buildGroupCompletion(g *config.Group) *complete.Command {
	sub := &complete.Command{
		Sub:   make(map[string]*complete.Command),
		Flags: make(map[string]complete.Predictor),
	}

	// Group-level args as flags
	for _, arg := range g.Args {
		addArgFlag(sub, arg)
	}

	for i := range g.Groups {
		sg := &g.Groups[i]
		sub.Sub[sg.Name] = buildGroupCompletion(sg)
	}
	for i := range g.Commands {
		c := &g.Commands[i]
		sub.Sub[c.Name] = buildCommandCompletion(c)
	}

	return sub
}

func buildCommandCompletion(c *config.Command) *complete.Command {
	cmd := &complete.Command{
		Flags: make(map[string]complete.Predictor),
	}

	// Command-level args as flags
	for _, arg := range c.Args {
		addArgFlag(cmd, arg)
	}

	// Positional args: allow anything (files, etc.)
	cmd.Args = predictAnything

	return cmd
}

func addGlobalFlags(cmd *complete.Command) {
	cmd.Flags["-h"] = predict.Nothing
	cmd.Flags["--help"] = predict.Nothing
	cmd.Flags["-v"] = predict.Nothing
	cmd.Flags["--verbose"] = predict.Nothing
	cmd.Flags["-V"] = predict.Nothing
	cmd.Flags["--version"] = predict.Nothing
	cmd.Flags["--dry-run"] = predict.Nothing
	cmd.Flags["--init"] = predict.Nothing
	cmd.Flags["--config"] = predict.Files("*")
}

func addArgFlag(cmd *complete.Command, arg config.Arg) {
	if arg.Short != "" {
		cmd.Flags["-"+arg.Short] = predict.Nothing
	}
	cmd.Flags["--"+arg.Name] = predict.Nothing
}

// anythingPredictor predicts nothing, allowing shell default completion.
type anythingPredictor struct{}

func (anythingPredictor) Predict(prefix string) []string {
	return nil
}

var predictAnything complete.Predictor = anythingPredictor{}

// completionScripts maps shell names to their completion scripts.
var completionScripts = map[string]string{
	"bash": "complete -C 'lota __complete' lota\n",
	"zsh": `#compdef lota
function _lota {
    local line="${LBUFFER}${RBUFFER}"
    local -a completions
    completions=($(env COMP_LINE="$line" COMP_POINT=${#LBUFFER} lota __complete))
    compadd -a completions
}
compdef _lota lota
`,
	"fish": `complete -c lota -f -a "(env COMP_LINE=(commandline) COMP_POINT=(commandline -C) lota __complete)"
`,
}

// GetCompletionScript returns the completion script content for a given shell.
func GetCompletionScript(shell string) (string, error) {
	script, ok := completionScripts[shell]
	if !ok {
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
	return script, nil
}

// PrintCompletionScript prints a shell completion installation script.
func PrintCompletionScript(shell string) error {
	script, err := GetCompletionScript(shell)
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}

// detectShell returns the shell name from the $SHELL environment variable.
func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	return filepath.Base(shell)
}

// installPath returns the standard installation path for a completion script.
func installPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	switch shell {
	case "bash":
		return filepath.Join(home, ".local", "share", "bash-completion", "completions", "lota"), nil
	case "zsh":
		return filepath.Join(home, ".config", "zsh", "completions", "_lota"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "completions", "lota.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// InstallCompletionScript writes the completion script to the standard location.
func InstallCompletionScript(shell string) error {
	if shell == "" {
		shell = detectShell()
		if shell == "" {
			return fmt.Errorf("unable to detect shell; specify explicitly: --install-completion bash|zsh|fish")
		}
	}

	script, err := GetCompletionScript(shell)
	if err != nil {
		return err
	}

	path, err := installPath(shell)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	fmt.Printf("Installed %s completion to %s\n", shell, path)
	return nil
}
