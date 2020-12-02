package rtd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var (
	dockerImpl docker2 = realDocker{}
)

type docker2 interface {
	run(ctx context.Context, stdoutBuf, stderrBuf *bytes.Buffer, args ...string) error
}

type realDocker struct{}

func (c realDocker) run(ctx context.Context, stdoutBuf, stderrBuf *bytes.Buffer, args ...string) error {
	stdoutBuf.Reset()
	stderrBuf.Reset()
	cmd := exec.CommandContext(ctx, "docker2", args...)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	log.Printf("running %s %s", "docker2", strings.Trim(fmt.Sprint(args), "[]"))
	err := cmd.Run()
	log.Printf("stdout\n%s", stdoutBuf.String())
	log.Printf("stderr\n%s", stderrBuf.String())
	return err
}
