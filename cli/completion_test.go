package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"lota/config"
)

func TestBuildCompletion_EmptyConfig(t *testing.T) {
	cfg := &config.AppConfig{}
	comp := BuildCompletion(cfg)

	if len(comp.Sub) != 0 {
		t.Errorf("expected 0 subcommands, got %d", len(comp.Sub))
	}

	expectedFlags := []string{"-h", "--help", "-v", "--verbose", "-V", "--version", "--dry-run", "--init", "--config"}
	for _, f := range expectedFlags {
		if _, ok := comp.Flags[f]; !ok {
			t.Errorf("expected global flag %q", f)
		}
	}
}

func TestBuildCompletion_WithGroupsAndCommands(t *testing.T) {
	cfg := &config.AppConfig{
		Groups: []config.Group{
			{
				Name: "db",
				Commands: []config.Command{
					{Name: "migrate"},
					{Name: "seed"},
				},
			},
		},
		Commands: []config.Command{
			{Name: "hello"},
			{Name: "world"},
		},
	}

	comp := BuildCompletion(cfg)

	if _, ok := comp.Sub["db"]; !ok {
		t.Error("expected 'db' group")
	}
	if _, ok := comp.Sub["hello"]; !ok {
		t.Error("expected 'hello' command")
	}
	if _, ok := comp.Sub["world"]; !ok {
		t.Error("expected 'world' command")
	}

	db := comp.Sub["db"]
	if _, ok := db.Sub["migrate"]; !ok {
		t.Error("expected 'db migrate' command")
	}
	if _, ok := db.Sub["seed"]; !ok {
		t.Error("expected 'db seed' command")
	}
}

func TestBuildCompletion_NestedGroups(t *testing.T) {
	cfg := &config.AppConfig{
		Groups: []config.Group{
			{
				Name: "infra",
				Groups: []config.Group{
					{
						Name: "db",
						Commands: []config.Command{
							{Name: "migrate"},
						},
					},
				},
			},
		},
	}

	comp := BuildCompletion(cfg)

	infra, ok := comp.Sub["infra"]
	if !ok {
		t.Fatal("expected 'infra' group")
	}

	db, ok := infra.Sub["db"]
	if !ok {
		t.Fatal("expected 'infra db' group")
	}

	if _, ok := db.Sub["migrate"]; !ok {
		t.Error("expected 'infra db migrate' command")
	}
}

func TestBuildCompletion_CommandFlags(t *testing.T) {
	cfg := &config.AppConfig{
		Commands: []config.Command{
			{
				Name: "build",
				Args: []config.Arg{
					{Name: "target", Short: "t"},
					{Name: "verbose", Short: "v"},
				},
			},
		},
	}

	comp := BuildCompletion(cfg)
	build := comp.Sub["build"]

	if _, ok := build.Flags["-t"]; !ok {
		t.Error("expected -t flag")
	}
	if _, ok := build.Flags["--target"]; !ok {
		t.Error("expected --target flag")
	}
	if _, ok := build.Flags["-v"]; !ok {
		t.Error("expected -v flag")
	}
	if _, ok := build.Flags["--verbose"]; !ok {
		t.Error("expected --verbose flag")
	}
}

func TestBuildCompletion_GroupFlags(t *testing.T) {
	cfg := &config.AppConfig{
		Groups: []config.Group{
			{
				Name: "deploy",
				Args: []config.Arg{
					{Name: "env", Short: "e"},
				},
				Commands: []config.Command{
					{Name: "prod"},
				},
			},
		},
	}

	comp := BuildCompletion(cfg)
	deploy := comp.Sub["deploy"]

	if _, ok := deploy.Flags["-e"]; !ok {
		t.Error("expected -e flag on group")
	}
	if _, ok := deploy.Flags["--env"]; !ok {
		t.Error("expected --env flag on group")
	}
}

func TestBuildCompletion_ConfigFlagPredictsFiles(t *testing.T) {
	cfg := &config.AppConfig{}
	comp := BuildCompletion(cfg)

	if comp.Flags["--config"] == nil {
		t.Error("expected --config to have a predictor")
	}
}

func TestPrintCompletionScript(t *testing.T) {
	tests := []struct {
		shell     string
		shouldErr bool
		contains  []string
		excludes  []string
	}{
		{
			shell:    "bash",
			contains: []string{"complete -C 'lota __complete' lota"},
		},
		{
			shell:    "zsh",
			contains: []string{"lota __complete"},
			excludes: []string{"export COMP_LINE", "export COMP_POINT", "local COMP_LINE", "local COMP_POINT"},
		},
		{
			shell:    "fish",
			contains: []string{"lota __complete"},
		},
		{
			shell:     "pwsh",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			// Capture stdout
			// PrintCompletionScript writes to stdout, so we redirect it temporarily
			// Using a pipe is the simplest approach
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe error: %v", err)
			}
			oldStdout := os.Stdout
			os.Stdout = w

			scriptErr := PrintCompletionScript(tt.shell)

			if err := w.Close(); err != nil {
				t.Errorf("failed to close pipe: %v", err)
			}
			os.Stdout = oldStdout

			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("read error: %v", err)
			}
			output := buf.String()

			if tt.shouldErr && scriptErr == nil {
				t.Error("expected error")
			}
			if !tt.shouldErr && scriptErr != nil {
				t.Errorf("unexpected error: %v", scriptErr)
			}
			if tt.shouldErr {
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(output, s) {
					t.Errorf("expected output to NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestGetCompletionScript(t *testing.T) {
	tests := []struct {
		shell     string
		shouldErr bool
		contains  string
	}{
		{"bash", false, "complete -C 'lota __complete' lota"},
		{"zsh", false, "lota __complete"},
		{"fish", false, "lota __complete"},
		{"pwsh", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script, err := GetCompletionScript(tt.shell)
			if tt.shouldErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !strings.Contains(script, tt.contains) {
				t.Errorf("expected script to contain %q, got:\n%s", tt.contains, script)
			}
		})
	}
}

func TestInstallCompletionScript(t *testing.T) {
	// Create a temporary home directory to avoid polluting the real one.
	tmpHome, err := os.MkdirTemp("", "lota-completion-test")
	if err != nil {
		t.Fatalf("mkdirtemp error: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpHome); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	// Override the install path by monkey-patching via a custom approach is hard;
	// instead we test the helper functions directly.
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			script, err := GetCompletionScript(shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if script == "" {
				t.Error("expected non-empty script")
			}
		})
	}

	// Test unsupported shell
	t.Run("unsupported", func(t *testing.T) {
		_, err := GetCompletionScript("pwsh")
		if err == nil {
			t.Error("expected error for unsupported shell")
		}
	})
}
