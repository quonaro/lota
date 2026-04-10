package cli

import (
	"fmt"
	"strings"
)

// GlobalFlags represents global CLI flags
type GlobalFlags struct {
	Help    bool
	Verbose bool
	Version bool
	DryRun  bool
	Init    bool
	Config  string
}

// ParseGlobalFlags parses global flags from CLI arguments.
// Returns flags and remaining arguments (excluding global flags).
// Only flags that appear before the first non-flag token (command name) are
// treated as global flags. Everything from the first non-flag token onward
// (including any flags mixed in after it) is returned as remaining args.
func ParseGlobalFlags(args []string) (GlobalFlags, []string) {
	flags := GlobalFlags{}
	i := 0

	for i < len(args) {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			break
		}

		known := true
		switch arg {
		case "--help", "-h":
			flags.Help = true
		case "--verbose", "-v":
			flags.Verbose = true
		case "--version", "-V":
			flags.Version = true
		case "--dry-run":
			flags.DryRun = true
		case "--init":
			flags.Init = true
		case "--config":
			if i+1 < len(args) {
				i++
				flags.Config = args[i]
			}
		default:
			known = false
		}

		if !known {
			break
		}
		i++
	}

	return flags, args[i:]
}

// HandleGlobalFlags handles global flags.
// Returns true if program should exit after handling.
func HandleGlobalFlags(flags GlobalFlags) bool {
	if flags.Help {
		PrintHelp(flags.Config)
		return true
	}

	if flags.Version {
		PrintVersion()
		return true
	}

	if flags.Init {
		if err := InitConfig(flags.Config); err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
		return true
	}

	return false
}
