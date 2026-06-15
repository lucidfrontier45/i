package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var Verbose bool

func cmdOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "+ %s %s\n", name, strings.Join(args, " "))
	}
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
