package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chzyer/readline"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/kballard/go-shellquote"
	"github.com/posener/complete"

	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

const interactivePrompt = "\033[32m>\033[0m "

var _ readline.AutoCompleter = &FTLCompletion{}

type KongContextBinder func(ctx context.Context, kctx *kong.Context) context.Context

type exitPanic struct{}

type interactiveConsole struct {
	projectConfig projectconfig.Config
	l             *readline.Instance
	binder        KongContextBinder
	k             *kong.Kong
	closeWait     sync.WaitGroup
	closed        bool
}

func newInteractiveConsole(k *kong.Kong, binder KongContextBinder, projectConfig projectconfig.Config, eventSource *schemaeventsource.EventSource) (*interactiveConsole, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          interactivePrompt,
		InterruptPrompt: "^C",
		AutoComplete:    &FTLCompletion{app: k, view: eventSource.ViewOnly()},
		Listener: &ExitListener{cancel: func() {
			_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck // best effort
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("init readline: %w", err)
	}

	it := &interactiveConsole{k: k, binder: binder, l: l, projectConfig: projectConfig}
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

func RunInteractiveConsole(ctx context.Context, k *kong.Kong, binder KongContextBinder, projectConfig projectconfig.Config, eventSource *schemaeventsource.EventSource) error {
	if !readline.DefaultIsTerminal() {
		return nil
	}
	ic, err := newInteractiveConsole(k, binder, projectConfig, eventSource)
	if err != nil {
		return err
	}
	defer ic.Close()
	err = ic.run(ctx)
	if err != nil {
		return err
	}
	return nil

}

func (r *interactiveConsole) run(ctx context.Context) error {
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
		tsm.console = true
		tsm.consoleRefresh = r.l.Refresh
		tsm.recalculateLines()
		tsm.statusLock.Unlock()
	}
	context.AfterFunc(ctx, r.Close)
	l.CaptureExitSignal()
	// Overload the exit function to avoid exiting the process
	k.Exit = func(i int) {
		if i != 0 {
			_ = l.Close()
			_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck
		}
		// For a normal exit from an interactive command we need a special panic
		// we recover from this and continue the loop
		panic(exitPanic{})
	}

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
		args, err := shellquote.Split(line)
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
		if tsm != nil {
			if len(args) > 0 && args[0] == "goose" {
				tsm.consoleNewline("👤 " + strings.Join(args[1:], " "))
			} else {
				tsm.consoleNewline("> " + line)
			}
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(exitPanic); ok {
						return
					}
					panic(r)
				}
			}()
			// Dynamically update the kong app with language specific flags for the "ftl module new" command.
			languagePlugin, err := languageplugin.PrepareNewCmd(ctx, r.projectConfig, k, args)
			if err != nil {
				errorf("%s", err)
				return
			}
			kctx, err := k.Parse(args)
			if err != nil {
				errorf("%s", err)
				return
			}
			kctx.Bind(languagePlugin)
			subctx := r.binder(ctx, kctx)

			err = kctx.Run(subctx)
			if err != nil {
				errorf("error: %s", err)
				return
			}
		}()
	}
	_ = l.Close() //nolint:errcheck // best effort

	return nil
}

var _ readline.Listener = &ExitListener{}

type ExitListener struct {
	cancel func()
}

func (e ExitListener) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	if key == readline.CharInterrupt {
		e.cancel()
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
