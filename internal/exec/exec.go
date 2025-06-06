package exec

import (
	"bytes"
	"context"
	"os"
	"os/exec" //nolint:depguard
	"strings"
	"syscall"

	"github.com/alecthomas/errors"
	"github.com/kballard/go-shellquote"

	"github.com/block/ftl/common/log"
)

type Cmd struct {
	*exec.Cmd
	level log.Level
}

func LookPath(exe string) (string, error) {
	path, err := exec.LookPath(exe)
	return path, errors.WithStack(err)
}

func Capture(ctx context.Context, dir, exe string, args ...string) ([]byte, error) {
	cmd := Command(ctx, log.Debug, dir, exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	out, err := cmd.CombinedOutput()
	return out, errors.WithStack(err)
}

func Command(ctx context.Context, level log.Level, dir, exe string, args ...string) *Cmd {
	return CommandWithEnv(ctx, level, dir, []string{}, exe, args...)
}

func CommandWithEnv(ctx context.Context, level log.Level, dir string, env []string, exe string, args ...string) *Cmd {
	logger := log.FromContext(ctx)
	pgid, err := syscall.Getpgid(0)
	if err != nil {
		panic(err)
	}
	logger.Tracef("exec: cd %s && %s %s", shellquote.Join(dir), exe, shellquote.Join(args...))
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pgid:    pgid,
		Setpgid: true,
	}
	cmd.Dir = dir
	output := logger.WriterAt(level)
	cmd.Stdout = output
	cmd.Stderr = output
	return &Cmd{cmd, level}
}

// RunBuffered runs the command and captures the output. If the command fails, the output is logged.
func (c *Cmd) RunBuffered(ctx context.Context) error {
	outputBuffer := NewCircularBuffer(100)
	output := outputBuffer.WriterAt(ctx, c.level)
	c.Stdout = output
	c.Stderr = output

	err := c.Run()
	if err != nil {
		if ctx.Err() == nil {
			// Don't log on context cancellation
			log.FromContext(ctx).Errorf(err, "%s", outputBuffer.Bytes())
		}
		return errors.Wrap(err, shellquote.Join(c.Args...))
	}

	return nil
}

// RunStderrError runs the command and captures stderr. If the command fails, the stderr is returned as the error message.
func (c *Cmd) RunStderrError(ctx context.Context) error {
	errorBuffer := NewCircularBuffer(100)

	c.Stdout = nil
	c.Stderr = errorBuffer.WriterAt(ctx, c.level)

	if err := c.Run(); err != nil {
		return errors.Errorf("%s: %s", shellquote.Join(c.Args...), strings.TrimSpace(string(errorBuffer.Bytes())))
	}

	return nil
}

// Capture runs the command and captures the output. If the command fails, the stderr is returned as the error message.
func (c *Cmd) Capture(ctx context.Context) ([]byte, error) {
	outBuffer := &bytes.Buffer{}
	errorBuffer := NewCircularBuffer(100)

	c.Stdout = outBuffer
	c.Stderr = errorBuffer.WriterAt(ctx, c.level)

	if err := c.Run(); err != nil {
		return nil, errors.Errorf("%s: %s", shellquote.Join(c.Args...), strings.TrimSpace(string(errorBuffer.Bytes())))
	}

	return outBuffer.Bytes(), nil
}

// Kill sends a signal to the process group of the command.
func (c *Cmd) Kill(signal syscall.Signal) error {
	if c.Process == nil {
		return nil
	}
	return errors.WithStack(syscall.Kill(c.Process.Pid, signal))
}
