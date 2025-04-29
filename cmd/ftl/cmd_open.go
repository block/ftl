package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/devstate"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
)

const (
	autodetect = "auto"
	vscode     = "vscode"
	intellij   = "intellij"
)

type openCmd struct {
	Ref    reflection.Ref `arg:"" help:"Language of the module to create." placeholder:"MODULE.ITEM" predictor:"decls"`
	Editor string         `arg:"" help:"Editor to open the file with." enum:"auto,vscode,intellij" default:"auto"`

	TerminalProgram  string `help:"Helps detect if the terminal is running within an IDE, used for 'auto' IDE detection" env:"TERM_PROGRAM" hidden:""`
	TerminalEmulator string `help:"Helps detect if the terminal is running within an IDE, used for 'auto' IDE detection" env:"TERMINAL_EMULATOR" hidden:""`
}

func (i openCmd) Run(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient adminpbconnect.AdminServiceClient, pc projectconfig.Config) error {
	// Currently dev state is the easiest way to get schema and module paths
	// We should use admin client to get schema directly when we have a single directory for modules in a project
	state, err := devstate.WaitForDevState(ctx, buildEngineClient, adminClient, false)
	if err != nil {
		return errors.Wrapf(err, "could not get dev state")
	}

	odecl, _ := state.Schema.ResolveWithModule(&schema.Ref{Module: i.Ref.Module, Name: i.Ref.Name})
	decl, ok := odecl.Get()
	if !ok {
		return fmt.Errorf("could not find %q", i.Ref)
	}
	if decl.Position().Filename == "" {
		return fmt.Errorf("could not find file of %q", i.Ref)
	}

	if i.Editor == autodetect {
		if i.TerminalProgram == "vscode" {
			i.Editor = vscode
		} else if strings.Contains(i.TerminalEmulator, "JetBrains") {
			i.Editor = intellij
		} else {
			return fmt.Errorf("could not auto choose default editor, expected one of %s", strings.Join([]string{vscode, intellij}, ", "))
		}
	}
	switch i.Editor {
	case vscode:
		return openVisualStudioCode(ctx, decl.Position(), pc.Root())
	case intellij:
		return openIntelliJ(ctx, decl.Position(), pc.Root())
	default:
		return fmt.Errorf("unsupported editor %q, expected one of %s", i.Editor, strings.Join([]string{vscode, intellij}, ", "))
	}
}

func openVisualStudioCode(ctx context.Context, pos schema.Position, projectRoot string) error {
	path := pos.Filename + ":" + fmt.Sprint(pos.Line)
	if pos.Column > 0 {
		path += ":" + fmt.Sprint(pos.Column)
	}
	err := exec.Command(ctx, log.Debug, ".", "code", projectRoot, "--goto", path).RunBuffered(ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open visual studio code")
	}
	return nil
}

func openIntelliJ(ctx context.Context, pos schema.Position, projectRoot string) error {
	err := exec.Command(ctx, log.Debug, ".", "idea", projectRoot, "--line", strconv.Itoa(pos.Line), "--column", strconv.Itoa(pos.Column), pos.Filename).RunBuffered(ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open IntelliJ IDEA")
	}
	return nil
}
