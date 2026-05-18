package cli

import (
	"lota/config"
	"testing"
)

func TestIsFlagArg(t *testing.T) {
	tests := []struct {
		name     string
		arg      config.Arg
		expected bool
	}{
		{
			name:     "plain positional",
			arg:      config.Arg{Name: "filename", Type: "str"},
			expected: false,
		},
		{
			name:     "positional with type int",
			arg:      config.Arg{Name: "count", Type: "int"},
			expected: false,
		},
		{
			name:     "short alias makes flag",
			arg:      config.Arg{Name: "output", Short: "o", Type: "str"},
			expected: true,
		},
		{
			name:     "bool type makes flag",
			arg:      config.Arg{Name: "verbose", Type: "bool"},
			expected: true,
		},
		{
			name:     "default value is still positional",
			arg:      config.Arg{Name: "env", Type: "str", Default: "dev"},
			expected: false,
		},
		{
			name:     "wildcard is positional",
			arg:      config.Arg{Name: "cmd", Wildcard: true},
			expected: false,
		},
		{
			name:     "wildcard with default is still positional",
			arg:      config.Arg{Name: "cmd", Wildcard: true, Default: "default"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFlagArg(tt.arg)
			if got != tt.expected {
				t.Errorf("isFlagArg(%+v) = %v, want %v", tt.arg, got, tt.expected)
			}
		})
	}
}

func TestBuildCommandUsage(t *testing.T) {
	positional := []config.Arg{
		{Name: "service", Required: true, Type: "str"},
		{Name: "env", Required: false, Type: "str"},
		{Name: "files", Type: "arr"},
		{Name: "tail", Wildcard: true},
	}
	flags := []config.Arg{{Name: "verbose", Type: "bool"}}

	got := buildCommandUsage("infra deploy", positional, flags)
	want := "Usage: lota infra deploy [OPTIONS] <SERVICE> [<ENV>] [<FILES...>] [...<TAIL>]"
	if got != want {
		t.Fatalf("buildCommandUsage() = %q, want %q", got, want)
	}
}

func TestUsageArgName(t *testing.T) {
	if got := usageArgName("service-name"); got != "SERVICE_NAME" {
		t.Fatalf("usageArgName() = %q, want %q", got, "SERVICE_NAME")
	}
	if got := usageArgName("!!!"); got != "ARG" {
		t.Fatalf("usageArgName() = %q, want %q", got, "ARG")
	}
}

func TestSeparateArgs(t *testing.T) {
	args := []config.Arg{
		{Name: "filename", Type: "str"},            // positional
		{Name: "output", Short: "o", Type: "str"},  // flag
		{Name: "verbose", Type: "bool"},            // flag
		{Name: "env", Type: "str", Default: "dev"}, // positional (has default)
		{Name: "cmd", Wildcard: true},              // positional
	}

	positional, flags := separateArgs(args)

	if len(positional) != 3 {
		t.Errorf("expected 3 positional args, got %d", len(positional))
	}
	if len(flags) != 2 {
		t.Errorf("expected 2 flag args, got %d", len(flags))
	}

	expectedPositional := []string{"filename", "env", "cmd"}
	for i, name := range expectedPositional {
		if positional[i].Name != name {
			t.Errorf("positional[%d].Name = %q, want %q", i, positional[i].Name, name)
		}
	}

	expectedFlags := []string{"output", "verbose"}
	for i, name := range expectedFlags {
		if flags[i].Name != name {
			t.Errorf("flags[%d].Name = %q, want %q", i, flags[i].Name, name)
		}
	}
}
