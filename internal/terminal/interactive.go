package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chzyer/readline"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/kballard/go-shellquote"
	"github.com/posener/complete"

	"github.com/block/ftl/internal/schema/schemaeventsource"
)

const interactivePrompt = "\033[1;32m❯\033[0m "

var _ readline.AutoCompleter = &FTLCompletion{}

var gooseCmdRegex = regexp.MustCompile(`^(ftl\s+)?goose\s+(.*)`)

type interactiveConsole struct {
	l                   *readline.Instance
	k                   *kong.Kong
	closeWait           sync.WaitGroup
	closed              bool
	currentLineCallback func(string)
}
type CommandExecutor func(ctx context.Context, k *kong.Kong, args []string, additionalExit func(int)) error

func newInteractiveConsole(k *kong.Kong, eventSource *schemaeventsource.EventSource) (*interactiveConsole, error) {
	it := &interactiveConsole{k: k}
	l, err := readline.NewEx(&readline.Config{
		Prompt:          interactivePrompt,
		InterruptPrompt: "^C",
		AutoComplete:    &FTLCompletion{app: k, view: eventSource.ViewOnly()},
		Listener:        it,
	})
	it.l = l
	if err != nil {
		return nil, fmt.Errorf("init readline: %w", err)
	}

	it.closeWait.Add(1)
	return it, nil
}

func (r *interactiveConsole) Close() {
	if r.closed {
		return
	}
	r.closed = true
	err := r.l.Close()
	if err != nil {
		return
	}
	r.closeWait.Wait()
}

func RunInteractiveConsole(ctx context.Context, k *kong.Kong, eventSource *schemaeventsource.EventSource, executor CommandExecutor) error {
	if !readline.DefaultIsTerminal() {
		return nil
	}
	ic, err := newInteractiveConsole(k, eventSource)
	if err != nil {
		return err
	}
	defer ic.Close()
	err = ic.run(ctx, executor)
	if err != nil {
		return err
	}
	return nil
}

func (r *interactiveConsole) run(ctx context.Context, executor CommandExecutor) error {
	if !readline.DefaultIsTerminal() {
		return nil
	}
	l := r.l
	k := r.k
	defer r.closeWait.Done()

	sm := FromContext(ctx)
	var tsm *terminalStatusManager
	ok := false
	if tsm, ok = sm.(*terminalStatusManager); ok {
		tsm.statusLock.Lock()
		tsm.clearStatusMessages()
		tsm.consoleRefresh = r.l.Refresh
		tsm.recalculateLines()
		tsm.statusLock.Unlock()
	}
	context.AfterFunc(ctx, r.Close)
	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if errors.Is(err, readline.ErrInterrupt) {

			if len(line) == 0 {
				break
			}
			continue
		} else if errors.Is(err, io.EOF) {
			defer func() { // We want the earlier defer to run first
				if !r.closed {
					// We only call sigint if this closure was because of ctrl+D
					_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck
				}
			}()
			return nil
		}
		if tsm != nil {
			tsm.consoleNewline(line)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var args []string
		if gooseCmdRegex.MatchString(line) {
			// Do not parse goose command text as a normal shell command. It is conversational and so apostrophes are fine and quotation marks should be kept.
			args = []string{
				"goose",
				gooseCmdRegex.FindStringSubmatch(line)[2],
			}
		} else {
			args, err = shellquote.Split(line)
			if err != nil {
				errorf("%s", err)
				continue
			}
			if len(args) > 0 && args[0] == "ftl" {
				args = args[1:]
			}
			if len(args) == 0 {
				continue
			}
		}
		if tsm != nil {
			if len(args) > 0 && args[0] == "goose" {
				tsm.consoleNewline("👤 " + strings.Join(args[1:], " "))
			} else {
				tsm.consoleNewline("> " + line)
			}
		}
		if err := executor(ctx, k, args, func(i int) {
			_ = l.Close()
		}); err != nil {
			errorf("%s", err)
			continue
		}
	}
	_ = l.Close() //nolint:errcheck // best effort

	return nil
}

func (r *interactiveConsole) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	if key == readline.CharInterrupt {
		_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck // best effort
	}
	if r.currentLineCallback != nil {
		r.currentLineCallback(string(line))
	}
	return line, pos, true
}

func errorf(format string, args ...any) {
	fmt.Printf("\033[31m%s\033[0m\n", fmt.Sprintf(format, args...))
}

type FTLCompletion struct {
	app  *kong.Kong
	view *schemaeventsource.View
}

func (f *FTLCompletion) Do(line []rune, pos int) ([][]rune, int) {
	parser := f.app
	if parser == nil {
		return nil, 0
	}
	all := []string{}
	completed := []string{}
	last := ""
	lastCompleted := ""
	lastSpace := false
	// We don't care about anything past pos
	// this completer can't handle completing in the middle of things
	if pos < len(line) {
		line = line[:pos]
	}
	current := 0
	for i, arg := range line {
		if i == pos {
			break
		}
		if arg == ' ' {
			lastWord := string(line[current:i])
			all = append(all, lastWord)
			completed = append(completed, lastWord)
			current = i + 1
			lastSpace = true
		} else {
			lastSpace = false
		}
	}
	if pos > 0 {
		if lastSpace {
			lastCompleted = all[len(all)-1]
		} else {
			if current < len(line) {
				last = string(line[current:])
				all = append(all, last)
			}
			if len(all) > 0 {
				lastCompleted = all[len(all)-1]
			}
		}
	}

	args := complete.Args{
		Completed:     completed,
		All:           all,
		Last:          last,
		LastCompleted: lastCompleted,
	}

	command, err := kongcompletion.Command(parser, kongcompletion.WithPredictors(Predictors(f.view)))
	if err != nil {
		// TODO handle error
		println(err.Error())
	}
	result := command.Predict(args)
	runes := [][]rune{}
	for _, s := range result {
		if !strings.HasPrefix(s, last) || s == "interactive" {
			continue
		}
		s = s[len(last):]
		str := []rune(s)
		runes = append(runes, str)
	}
	return runes, pos
}
