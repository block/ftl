package editor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

const (
	VSCode   = "vscode"
	IntelliJ = "intellij"
	Cursor   = "cursor"
	Zed      = "zed"
)

var SupportedEditors = []string{VSCode, IntelliJ, Cursor, Zed}

// OpenFileInEditor opens the given file path at the specified position in the selected editor.
func OpenFileInEditor(ctx context.Context, editor string, pos schema.Position, projectRoot string) error {
	executable, args, workDir, err := getEditorCmdArgs(editor, pos, projectRoot)
	if err != nil {
		return err
	}

	err = exec.Command(ctx, log.Debug, workDir, executable, args...).RunBuffered(ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open file in %s: ensure the '%s' shell command is installed and configured correctly", editor, executable)
	}
	return nil
}

func getEditorCmdArgs(editor string, pos schema.Position, projectRoot string) (executable string, args []string, workDir string, err error) {
	switch editor {
	case VSCode:
		executable = "code"
		path := formatPathWithPosition(pos, true)
		args = []string{projectRoot, "--goto", path}
		workDir = "."

	case IntelliJ:
		executable = "idea"
		args = []string{projectRoot, "--line", strconv.Itoa(pos.Line), "--column", strconv.Itoa(pos.Column), pos.Filename}
		workDir = "."

	case Cursor:
		executable = "cursor"
		path := formatPathWithPosition(pos, true)
		args = []string{projectRoot, "--goto", path}
		workDir = "."

	case Zed:
		executable = "zed"
		path := formatPathWithPosition(pos, true)
		args = []string{path}
		workDir = projectRoot

	default:
		err = errors.Errorf("unsupported editor %q, expected one of %s", editor, strings.Join(SupportedEditors, ", "))
	}
	return
}

func formatPathWithPosition(pos schema.Position, includeColumn bool) string {
	path := pos.Filename + ":" + fmt.Sprint(pos.Line)
	if includeColumn && pos.Column > 0 {
		path += ":" + fmt.Sprint(pos.Column)
	}
	return path
}
