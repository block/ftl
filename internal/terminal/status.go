package terminal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/tidwall/pretty"
	"golang.org/x/term"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

const ansiUpOneLine = "\u001B[1A"
const ansiClearLine = "\u001B[2K"
const ansiResetTextColor = "\u001B[39m"

type BuildState string

const BuildStateWaiting BuildState = "Waiting"
const BuildStateBuilding BuildState = "Building"
const BuildStateBuilt BuildState = "Built"
const BuildStateDeployWaiting BuildState = "DeployWaiting"
const BuildStateDeploying BuildState = "Deploying"
const BuildStateDeployed BuildState = "Deployed"
const BuildStateFailed BuildState = "Failed"
const BuildStateTerminated BuildState = "Terminated"

// moduleStatusPadding is the padding between module status entries
// it accounts for the icon, the module name, and the padding between them
const moduleStatusPadding = 5

var _ StatusManager = &terminalStatusManager{}
var _ StatusLine = &terminalStatusLine{}

var buildColors map[BuildState]string
var buildStateIcon map[BuildState]func(int) string
var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func init() {
	buildColors = map[BuildState]string{
		BuildStateWaiting:       "\u001B[38;5;244m", // Dimmed gray for waiting
		BuildStateBuilding:      "\u001B[38;5;33m",  // Bright blue for building
		BuildStateBuilt:         "\u001B[38;5;33m",  // Bright blue for success
		BuildStateDeployWaiting: "\u001B[38;5;244m", // Dimmed gray for waiting
		BuildStateDeploying:     "\u001B[38;5;40m",  // Bright green for deploying
		BuildStateDeployed:      "\u001B[38;5;40m",  // Bright green for success
		BuildStateFailed:        "\u001B[38;5;196m", // Bright red for failure
	}

	waiting := func(int) string {
		return "◌"
	}
	spin := func(spinnerCount int) string {
		return spinnerChars[spinnerCount%len(spinnerChars)]
	}
	success := func(int) string {
		return "●"
	}
	failed := func(int) string {
		return "✖"
	}

	buildStateIcon = map[BuildState]func(int) string{
		BuildStateWaiting:       waiting,
		BuildStateBuilding:      spin,
		BuildStateBuilt:         success,
		BuildStateDeployWaiting: waiting,
		BuildStateDeploying:     spin,
		BuildStateDeployed:      success,
		BuildStateFailed:        failed,
	}
}

type StatusManager interface {
	Close()
	NewStatus(message string) StatusLine
	NewDecoratedStatus(prefix string, suffix string, message string) StatusLine
	IntoContext(ctx context.Context) context.Context
	SetModuleState(module string, state BuildState)
}

type StatusLine interface {
	SetMessage(message string)
	Close()
}

type terminalStatusManager struct {
	old                *os.File
	oldErr             *os.File
	read               *os.File
	write              *os.File
	closed             atomic.Value[bool]
	totalStatusLines   int
	statusLock         sync.Mutex
	lines              []*terminalStatusLine
	moduleLine         *terminalStatusLine
	moduleStates       map[string]BuildState
	height             int
	width              int
	exitWait           sync.WaitGroup
	consoleRefresh     func()
	spinnerCount       int
	interactiveConsole optional.Option[*interactiveConsole]
	interactiveLines   int
}

type statusKey struct{}

var statusKeyInstance = statusKey{}

func NewStatusManager(ctx context.Context) StatusManager {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return &noopStatusManager{}
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return &noopStatusManager{}
	}
	sm := &terminalStatusManager{statusLock: sync.Mutex{}, moduleStates: map[string]BuildState{}, height: height, width: width, exitWait: sync.WaitGroup{}}
	sm.exitWait.Add(1)
	sm.old = os.Stdout
	sm.oldErr = os.Stderr
	sm.read, sm.write, err = os.Pipe()

	if err != nil {
		return &noopStatusManager{}
	}
	os.Stdout = sm.write
	os.Stderr = sm.write

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGWINCH)
		for range c {
			sm.handleResize()
		}
	}()

	go func() {
		current := ""
		closed := false
		for {
			buf := bytes.Buffer{}
			rawData := make([]byte, 104)
			n, err := sm.read.Read(rawData)
			if err != nil {
				if current != "" {
					sm.writeLine(current, true)
				}
				if !closed {
					sm.statusLock.Lock()
					sm.clearStatusMessages()
					sm.exitWait.Done()
					sm.statusLock.Unlock()
				}
				return
			}
			if closed {
				// When we are closed we just write the data to the old stdout
				_, _ = sm.old.Write(rawData[:n]) //nolint:errcheck
				continue
			}
			buf.Write(rawData[:n])
			for buf.Len() > 0 {
				d, s, err := buf.ReadRune()
				if d == 0 {
					// Null byte, we are done
					// we keep running though as there may be more data on exit
					// that we handle on a best effort basis
					if current != "" {
						sm.writeLine(current, true)
					}
					if !closed {
						sm.statusLock.Lock()
						sm.exitWait.Done()
						closed = true
						sm.statusLock.Unlock()
					}
					continue
				}
				if err != nil {
					// EOF, need to read more data
					break
				}
				if d == utf8.RuneError && s == 1 {
					if buf.Available() < 4 && !sm.closed.Load() {
						_ = buf.UnreadByte() //nolint:errcheck
						// Need to read more data, probably not a full rune
						break
					}
					// Otherwise, ignore the error, not much we can do
					continue
				}
				if d == '\n' {
					if current != "" {
						sm.writeLine(current, false)
						current = ""
					}
				} else {
					current += string(d)
				}
			}
		}
	}()

	// Animate the spinners
	go func() {
		for !sm.closed.Load() {
			time.Sleep(150 * time.Millisecond)
			sm.statusLock.Lock()
			if sm.spinnerCount == len(spinnerChars)-1 {
				sm.spinnerCount = 0
			} else {
				sm.spinnerCount++
			}
			// only redraw if not stable
			stable := true
			for _, state := range sm.moduleStates {
				if state != BuildStateDeployed && state != BuildStateBuilt {
					stable = false
					break
				}
			}
			if !stable {
				sm.recalculateLines()
			}
			sm.statusLock.Unlock()
		}
	}()

	return sm
}

func UpdateModuleState(ctx context.Context, module string, state BuildState) {
	sm := FromContext(ctx)
	sm.SetModuleState(module, state)
}

// PrintJSON prints a json string to the terminal
// It probably doesn't belong here, but it will be moved later with the interactive terminal work
func PrintJSON(ctx context.Context, json []byte) {
	sm := FromContext(ctx)
	if _, ok := sm.(*terminalStatusManager); ok {
		// ANSI enabled
		fmt.Printf("%s\n", pretty.Color(pretty.Pretty(json), nil))
	} else {
		fmt.Printf("%s\n", json)
	}
}

func IsANSITerminal(ctx context.Context) bool {
	sm := FromContext(ctx)
	_, ok := sm.(*terminalStatusManager)
	return ok
}

func (r *terminalStatusManager) clearStatusMessages() {
	if r.closed.Load() {
		return
	}
	if r.statusLock.TryLock() {
		panic("clearStatusMessages called without holding the lock")
	}
	if r.totalStatusLines == 0 {
		return
	}
	count := r.totalStatusLines
	if r.interactiveConsole.Ok() {
		// With the interactive console the cursor sits on the last line, so we just clear, we don't move up
		r.underlyingWrite(ansiClearLine)
	} else if count > 0 {
		// Without the interactive console, only move up if we have status lines to clear
		r.underlyingWrite(ansiUpOneLine + ansiClearLine)
	}
	for range count - 1 {
		r.underlyingWrite(ansiUpOneLine + ansiClearLine)
	}
}

func (r *terminalStatusManager) consoleNewline(line string) {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	count := r.totalStatusLines
	for range count {
		r.underlyingWrite(ansiUpOneLine + ansiClearLine)
	}
	r.underlyingWrite("\r" + line + "\n")
	r.redrawStatus()
}

func (r *terminalStatusManager) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, statusKeyInstance, r)
}

func FromContext(ctx context.Context) StatusManager {
	sm, ok := ctx.Value(statusKeyInstance).(StatusManager)
	if !ok {
		return &noopStatusManager{}
	}
	return sm
}

func (r *terminalStatusManager) NewStatus(message string) StatusLine {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	line := &terminalStatusLine{manager: r, message: message}
	r.newStatusInternal(line)
	return line
}
func (r *terminalStatusManager) NewDecoratedStatus(prefix string, suffix string, message string) StatusLine {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	line := &terminalStatusLine{manager: r, message: message, prefix: prefix, suffix: suffix}
	r.newStatusInternal(line)
	return line
}
func (r *terminalStatusManager) newStatusInternal(line *terminalStatusLine) {
	if r.closed.Load() {
		return
	}
	for i, l := range r.lines {
		if l.priority < line.priority {
			r.lines = slices.Insert(r.lines, i, line)
			r.recalculateLines()
			return
		}
	}
	// If we get here we just append the line
	r.lines = append(r.lines, line)
	r.recalculateLines()
}

func (r *terminalStatusManager) SetModuleState(module string, state BuildState) {
	if module == "builtin" {
		return
	}
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	if state == BuildStateTerminated {
		delete(r.moduleStates, module)
	} else {
		r.moduleStates[module] = state
	}
	if r.moduleLine != nil {
		r.recalculateLines()
	} else {
		r.moduleLine = &terminalStatusLine{manager: r, priority: -10000}
		r.newStatusInternal(r.moduleLine)
	}
}

func (r *terminalStatusManager) Close() {
	r.statusLock.Lock()
	if r.closed.Load() {
		r.statusLock.Unlock()
		return
	}
	if it, ok := r.interactiveConsole.Get(); ok {
		it.Close()
	}
	r.clearStatusMessages()
	r.closed.Store(true)
	r.totalStatusLines = 0
	r.lines = []*terminalStatusLine{}
	os.Stdout = r.old // restoring the real stdout
	os.Stderr = r.oldErr
	// We send a null byte to the write pipe to unblock the read
	_, _ = r.write.Write([]byte{0}) //nolint:errcheck
	r.statusLock.Unlock()
	r.exitWait.Wait()
}

func (r *terminalStatusManager) writeLine(s string, last bool) {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	if r.closed.Load() {
		os.Stdout.WriteString(s + "\n") //nolint:errcheck
		return
	}
	if !last {
		s += "\n"
	}

	if r.totalStatusLines == 0 && r.interactiveConsole.Ok() {
		// In interactive mode, we need to preserve the command line
		// by moving up one line before writing the output
		r.underlyingWrite(ansiUpOneLine + ansiClearLine + s)
		return
	}
	r.clearStatusMessages()
	r.underlyingWrite("\r" + s)
	if !last {
		r.redrawStatus()
	}
}
func (r *terminalStatusManager) redrawStatus() {
	if r.statusLock.TryLock() {
		panic("redrawStatus called without holding the lock")
	}
	if r.totalStatusLines == 0 || r.closed.Load() {
		return
	}
	for i := len(r.lines) - 1; i >= 0; i-- {
		msg := r.lines[i].prefix + r.lines[i].message + r.lines[i].suffix
		if msg != "" {
			r.underlyingWrite("\r" + msg + "\n")
		}
	}
	if r.consoleRefresh != nil {
		// If there is more than one line we need to actually make space for it to render
		for range r.interactiveLines - 1 {
			r.underlyingWrite("\n")
		}
		r.consoleRefresh()
	}
}

func truncateModuleName(name string, maxLength int) string {
	if len(name) <= maxLength {
		return name
	}
	if maxLength < 5 {
		return name[:maxLength]
	}

	half := (maxLength - 3) / 2
	return name[:half] + "..." + name[len(name)-half:]
}

func (r *terminalStatusManager) recalculateLines() {
	r.clearStatusMessages()
	total := 0
	if len(r.moduleStates) > 0 && r.moduleLine != nil {
		total++
		// Calculate the longest module name to ensure consistent spacing
		maxLength := 0
		keys := make([]string, 0, len(r.moduleStates))
		for k := range r.moduleStates {
			if len(k) > maxLength {
				maxLength = len(k)
			}
			keys = append(keys, k)
		}

		// Cap the maximum length to 25% of terminal width
		maxAllowedLength := r.width / 4
		if maxLength > maxAllowedLength {
			maxLength = maxAllowedLength
		}

		entryLength := maxLength + moduleStatusPadding

		msg := ""
		perLine := r.width / entryLength
		if perLine == 0 {
			perLine = 1
		}
		slices.Sort(keys)
		multiLine := false
		for i, k := range keys {
			if i%perLine == 0 && i > 0 {
				msg += "\n"
				multiLine = true
				total++
			}
			state := r.moduleStates[k]
			icon := buildStateIcon[state](r.spinnerCount)

			displayName := truncateModuleName(k, maxLength)
			moduleText := buildColors[state] + icon + " " + log.ScopeColor(k) + displayName + buildColors[state] + ansiResetTextColor

			padLength := entryLength - len(displayName) - 2 // -2 for icon and space
			pad := strings.Repeat(" ", padLength)
			msg += moduleText + pad
		}
		if !multiLine {
			msg = strings.TrimSpace(msg)
		}
		r.moduleLine.message = msg
	} else if r.moduleLine != nil {
		r.moduleLine.message = ""
	}
	for _, i := range r.lines {
		if i.message != "" && i != r.moduleLine {
			total++
			total += countLines(i.prefix+i.message+i.suffix, r.width)
		}
	}
	total += r.interactiveLines

	r.totalStatusLines = total

	r.redrawStatus()
}

func (r *terminalStatusManager) underlyingWrite(messages string) {
	_, _ = r.old.WriteString(messages) //nolint:errcheck
}

func countLines(s string, width int) int {
	if s == "" {
		return 0
	}
	lines := 0
	curLength := 0
	inANSI := false
	for _, r := range s {
		if inANSI {
			if r == 'm' {
				inANSI = false
			}
			continue
		}
		if r == '\u001B' { // Check for ANSI escape sequence start
			inANSI = true
			continue
		}
		if r == '\n' {
			lines++
			curLength = 0
		} else {
			// Assume most runes have width 1 for simplicity here.
			// TODO: count unicode characters properly, probably using a library
			curLength++
			if curLength > width {
				lines++
				curLength = 1
			}
		}
	}
	return lines
}

type terminalStatusLine struct {
	manager  *terminalStatusManager
	message  string
	priority int
	prefix   string
	suffix   string
}

func (r *terminalStatusLine) Close() {
	r.manager.statusLock.Lock()
	defer r.manager.statusLock.Unlock()
	for i := range r.manager.lines {
		if r.manager.lines[i] == r {
			r.manager.lines = append(r.manager.lines[:i], r.manager.lines[i+1:]...)
			r.manager.recalculateLines()
			return
		}
	}
	r.manager.redrawStatus()
}

func (r *terminalStatusLine) SetMessage(message string) {
	r.manager.statusLock.Lock()
	defer r.manager.statusLock.Unlock()
	r.message = message
	r.manager.recalculateLines()
}

func LaunchEmbeddedConsole(ctx context.Context, k *kong.Kong, eventSource *schemaeventsource.EventSource, executor CommandExecutor) {
	sm := FromContext(ctx)
	if tsm, ok := sm.(*terminalStatusManager); ok {
		it, err := newInteractiveConsole(k, eventSource)
		if err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		it.currentLineCallback = func(s string) {
			lines := countLines("> "+s, tsm.width) + 1
			if lines != tsm.interactiveLines {
				tsm.statusLock.Lock()
				defer tsm.statusLock.Unlock()
				if lines > tsm.interactiveLines {
					// The newline has already rendered so we need to add the lines to the total
					// This is so that it actually gets cleared
					tsm.totalStatusLines += lines - tsm.interactiveLines
				}
				tsm.interactiveLines = lines
				tsm.recalculateLines()
			}
		}
		tsm.interactiveLines = 1
		tsm.interactiveConsole = optional.Some(it)
		go func() {
			err := it.run(ctx, executor)
			if err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n", err)
				return
			}
		}()
	}
}

var _ io.Writer = &statusLineWriter{}

func StatusLineAsWriter(status StatusLine) io.Writer {
	return &statusLineWriter{status: status}
}

type statusLineWriter struct {
	status StatusLine
	line   string
}

func (s *statusLineWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if c == '\n' {
			s.status.SetMessage(s.line)
			s.line = ""
		} else {
			s.line += string(c)
		}
	}
	return len(p), nil
}

func (r *terminalStatusManager) handleResize() {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	r.width, r.height, _ = term.GetSize(int(r.old.Fd())) //nolint:errcheck
	r.recalculateLines()
}
