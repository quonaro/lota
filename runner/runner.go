package runner

import (
	"fmt"
	"lota/config"
	"os"
	"os/exec"
)

// RunOptions controls execution behavior
type RunOptions struct {
	Verbose bool
	DryRun  bool
}

// executeShell runs a script in shell with environment variables
func executeShell(script string, env []string, shell string) error {
	cmd := exec.Command(shell, script)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecuteCommand(cmd *config.Command, context InterpolationContext, opts RunOptions, shell string) error {
	env := VarsToEnv(context.Vars)

	if opts.Verbose {
		fmt.Printf("[verbose] command: %s\n", cmd.Name)
		fmt.Println("[verbose] vars:")
		for k, v := range context.Vars {
			fmt.Printf("  %s=%s\n", k, v)
		}
		fmt.Println("[verbose] args:")
		for k, v := range context.Args {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	// before hook
	if cmd.Before != "" {
		interpolatedBefore, err := Interpolate(cmd.Before, context)
		if err != nil {
			return fmt.Errorf("before hook interpolation failed: %w", err)
		}
		if opts.Verbose {
			fmt.Printf("[verbose] before: %s\n", interpolatedBefore)
		}
		if opts.DryRun {
			fmt.Printf("[dry-run] before:\n%s\n", interpolatedBefore)
		} else {
			if err := executeShell(interpolatedBefore, env, shell); err != nil {
				return fmt.Errorf("before hook failed: %w", err)
			}
		}
	}

	// after hook via defer (always executes)
	defer func() {
		if cmd.After != "" {
			interpolatedAfter, err := Interpolate(cmd.After, context)
			if err != nil {
				fmt.Printf("after hook interpolation failed: %v\n", err)
				return
			}
			if opts.Verbose {
				fmt.Printf("[verbose] after: %s\n", interpolatedAfter)
			}
			if opts.DryRun {
				fmt.Printf("[dry-run] after:\n%s\n", interpolatedAfter)
				return
			}
			if err := executeShell(interpolatedAfter, env, shell); err != nil {
				fmt.Printf("after hook failed: %v\n", err)
			}
		}
	}()

	// script
	if cmd.Script != "" {
		interpolatedScript, err := Interpolate(cmd.Script, context)
		if err != nil {
			return err
		}
		if opts.Verbose {
			fmt.Printf("[verbose] script: %s\n", interpolatedScript)
		}
		if opts.DryRun {
			fmt.Printf("[dry-run] script:\n%s\n", interpolatedScript)
			return nil
		}
		return executeShell(interpolatedScript, env, shell)
	}

	return nil
}
