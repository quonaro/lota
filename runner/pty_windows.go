//go:build windows

package runner

import (
	"context"
	"io"
	"os/exec"
	"sync"
)

func runWithPTY(cmd *exec.Cmd, stdout, stderr io.Writer, ctx context.Context, shutdownOnce *sync.Once) (bool, error) {
	return false, nil
}
