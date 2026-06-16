package manager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var Verbose bool

func cmdOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "+ %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	if Verbose {
		cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
	} else {
		cmd.Stdout = &buf
		cmd.Stderr = &buf
	}
	err := cmd.Run()
	return buf.Bytes(), err
}
