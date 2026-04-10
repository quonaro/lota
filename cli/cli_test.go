package cli

import (
	"lota/config"
	"testing"
)

func TestParseCommandPath(t *testing.T) {
	tests := []struct {
		name         string
		input        []string
		expectedPath []string
		expectedArgs []string
	}{
		{
			name:         "empty input",
			input:        []string{},
			expectedPath: []string{},
			expectedArgs: []string{},
		},
		{
			name:         "single command",
			input:        []string{"build"},
			expectedPath: []string{"build"},
			expectedArgs: []string{},
		},
		{
			name:         "group and command",
			input:        []string{"dev", "run"},
			expectedPath: []string{"dev", "run"},
			expectedArgs: []string{},
		},
		{
			name:         "command with positional arg",
			input:        []string{"deploy", "production"},
			expectedPath: []string{"deploy", "production"},
			expectedArgs: []string{},
		},
		{
			name:         "command with flag stops at two elements",
			input:        []string{"build", "--verbose"},
			expectedPath: []string{"build"},
			expectedArgs: []string{"--verbose"},
		},
		{
			name:         "group and command with args",
			input:        []string{"dev", "run", "--port", "8080"},
			expectedPath: []string{"dev", "run"},
			expectedArgs: []string{"--port", "8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, args := ParseCommandPath(tt.input)
			if !stringSlicesEqual(path, tt.expectedPath) {
				t.Errorf("path = %v, want %v", path, tt.expectedPath)
			}
			if !stringSlicesEqual(args, tt.expectedArgs) {
				t.Errorf("args = %v, want %v", args, tt.expectedArgs)
			}
		})
	}
}

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectedFlags GlobalFlags
		expectedArgs  []string
		wantErr       bool
	}{
		{
			name:          "no flags",
			input:         []string{"command"},
			expectedFlags: GlobalFlags{},
			expectedArgs:  []string{"command"},
		},
		{
			name:          "verbose long flag",
			input:         []string{"--verbose", "command"},
			expectedFlags: GlobalFlags{Verbose: true},
			expectedArgs:  []string{"command"},
		},
		{
			name:          "verbose short flag",
			input:         []string{"-v", "command"},
			expectedFlags: GlobalFlags{Verbose: true},
			expectedArgs:  []string{"command"},
		},
		{
			name:          "help flag",
			input:         []string{"--help"},
			expectedFlags: GlobalFlags{Help: true},
			expectedArgs:  []string{},
		},
		{
			name:          "help short flag",
			input:         []string{"-h"},
			expectedFlags: GlobalFlags{Help: true},
			expectedArgs:  []string{},
		},
		{
			name:          "version flag",
			input:         []string{"-V"},
			expectedFlags: GlobalFlags{Version: true},
			expectedArgs:  []string{},
		},
		{
			name:          "dry-run flag",
			input:         []string{"--dry-run", "command"},
			expectedFlags: GlobalFlags{DryRun: true},
			expectedArgs:  []string{"command"},
		},
		{
			name:          "init flag",
			input:         []string{"--init"},
			expectedFlags: GlobalFlags{Init: true},
			expectedArgs:  []string{},
		},
		{
			name:          "config flag",
			input:         []string{"--config", "/path/to/config.yml", "command"},
			expectedFlags: GlobalFlags{Config: "/path/to/config.yml"},
			expectedArgs:  []string{"command"},
		},
		{
			name:          "unknown flag stops parsing",
			input:         []string{"--unknown", "--verbose"},
			expectedFlags: GlobalFlags{},
			expectedArgs:  []string{"--unknown", "--verbose"},
		},
		{
			name:          "multiple flags",
			input:         []string{"--verbose", "--dry-run", "command"},
			expectedFlags: GlobalFlags{Verbose: true, DryRun: true},
			expectedArgs:  []string{"command"},
		},
		{
			name:    "config flag without value",
			input:   []string{"--config"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseGlobalFlags(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if flags != tt.expectedFlags {
				t.Errorf("flags = %+v, want %+v", flags, tt.expectedFlags)
			}
			if !stringSlicesEqual(remaining, tt.expectedArgs) {
				t.Errorf("remaining = %v, want %v", remaining, tt.expectedArgs)
			}
		})
	}
}

func TestFindCommand(t *testing.T) {
	cfg := &config.AppConfig{
		Commands: []config.Command{
			{Name: "build", Script: "go build"},
			{Name: "test", Script: "go test ./..."},
		},
		Groups: []config.Group{
			{
				Name: "dev",
				Commands: []config.Command{
					{Name: "run", Script: "go run ."},
					{Name: "watch", Script: "air"},
				},
			},
		},
	}
	if err := cfg.BuildIndexes(); err != nil {
		t.Fatalf("BuildIndexes() error: %v", err)
	}

	tests := []struct {
		name          string
		path          []string
		expectExists  bool
		expectCommand bool
		expectGroup   bool
	}{
		{
			name:          "top-level command",
			path:          []string{"build"},
			expectExists:  true,
			expectCommand: true,
		},
		{
			name:          "another top-level command",
			path:          []string{"test"},
			expectExists:  true,
			expectCommand: true,
		},
		{
			name:         "non-existing command",
			path:         []string{"nonexistent"},
			expectExists: false,
		},
		{
			name:        "existing group",
			path:        []string{"dev"},
			expectExists: true,
			expectGroup: true,
		},
		{
			name:          "command inside group",
			path:          []string{"dev", "run"},
			expectExists:  true,
			expectCommand: true,
			expectGroup:   true,
		},
		{
			name:         "non-existing command in group",
			path:         []string{"dev", "nonexistent"},
			expectExists: false,
		},
		{
			name:         "non-existing group",
			path:         []string{"nonexistent", "run"},
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindCommand(cfg, tt.path)

			if result.Exists != tt.expectExists {
				t.Errorf("Exists = %v, want %v", result.Exists, tt.expectExists)
			}
			if tt.expectCommand && result.Command == nil {
				t.Error("expected Command to be non-nil")
			}
			if !tt.expectCommand && result.Command != nil {
				t.Errorf("expected Command to be nil, got %v", result.Command.Name)
			}
			if tt.expectGroup && result.Group == nil {
				t.Error("expected Group to be non-nil")
			}
		})
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
