package builderrors

import (
	"errors"
	"fmt"
	"sort"

	"github.com/alecthomas/types/optional"
)

type Position struct {
	Filename    string
	Offset      int
	StartColumn int
	EndColumn   int
	Line        int
}

func (p Position) String() string {
	columnStr := fmt.Sprintf("%d", p.StartColumn)
	if p.StartColumn != p.EndColumn {
		columnStr += fmt.Sprintf("-%d", p.EndColumn)
	}
	if p.Filename == "" {
		return fmt.Sprintf("%d:%s", p.Line, columnStr)
	}
	return fmt.Sprintf("%s:%d:%s", p.Filename, p.Line, columnStr)
}

type ErrorLevel int

const (
	UNDEFINED ErrorLevel = iota
	INFO
	WARN
	ERROR
)

type ErrorType int

const (
	UNSPECIFIED ErrorType = iota
	// FTL errors are errors that are generated by the FTL runtime or the language plugin.
	FTL
	// COMPILER errors are errors that are generated by the compiler itself.
	// LSP integration will not show these errors as they will be generated by the compiler.
	COMPILER
)

type Error struct {
	Type  ErrorType
	Msg   string
	Pos   optional.Option[Position]
	Level ErrorLevel
}

func (e Error) Error() string {
	if pos, ok := e.Pos.Get(); ok {
		return fmt.Sprintf("%s: %s", pos, e.Msg)
	}
	return e.Msg
}

func makeError(level ErrorLevel, pos Position, format string, args ...any) Error {
	return Error{Type: FTL, Msg: fmt.Sprintf(format, args...), Pos: optional.Some(pos), Level: level}
}

func Infof(pos Position, format string, args ...any) Error {
	return makeError(INFO, pos, format, args...)
}

func Warnf(pos Position, format string, args ...any) Error {
	return makeError(WARN, pos, format, args...)
}

func Errorf(pos Position, format string, args ...any) Error {
	return makeError(ERROR, pos, format, args...)
}

func Wrapf(pos Position, endColumn int, err error, format string, args ...any) Error {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	// Propagate existing error position if available
	var newPos Position
	if perr := (Error{}); errors.As(err, &perr) {
		if existingPos, ok := perr.Pos.Get(); ok {
			newPos = existingPos
		}
		args = append(args, perr.Msg)
	} else {
		newPos = pos
		args = append(args, err)
	}
	return makeError(ERROR, newPos, format, args...)
}

func SortErrorsByPosition(merr []Error) {
	if merr == nil {
		return
	}
	sort.Slice(merr, func(i, j int) bool {
		ipp := merr[i].Pos.Default(Position{})
		jpp := merr[j].Pos.Default(Position{})
		return ipp.Line < jpp.Line || (ipp.Line == jpp.Line && ipp.StartColumn < jpp.StartColumn) ||
			(ipp.Line == jpp.Line && ipp.StartColumn == jpp.StartColumn && ipp.EndColumn < jpp.EndColumn) ||
			(ipp.Line == jpp.Line && ipp.StartColumn == jpp.StartColumn && ipp == jpp && merr[i].Msg < merr[j].Msg)
	})
}

func ContainsTerminalError(errs []Error) bool {
	for _, e := range errs {
		if e.Level == ERROR {
			return true
		}
	}
	return false
}
