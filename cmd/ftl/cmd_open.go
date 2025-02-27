package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
)

type openCmd struct {
	Ref    reflection.Ref `arg:"" help:"Language of the module to create." placeholder:"MODULE.ITEM" predictor:"decls"`
	Editor string         `arg:"" help:"Editor to open the file with." enum:"auto,vscode,intellij" default:"auto"`

	TerminalProgram  string `help:"Terminal program this command is running in which can influence 'auto' editor" env:"TERM_PROGRAM" hidden:""`
	TerminalEmulator string `help:"Terminal emulator can influence 'auto' editor" env:"TERMINAL_EMULATOR" hidden:""`
}

func (i openCmd) Run(ctx context.Context, ktctx *kong.Context, client ftlv1connect.AdminServiceClient, pc projectconfig.Config) error {
	ref := i.Ref.ToSchema()
	resp, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("could not get schema: %w", err)
	}
	sch, err := mergedSchemaFromResp(resp.Msg)
	if err != nil {
		return err
	}
	decl, ok := sch.Resolve(ref).Get()
	if !ok {
		return fmt.Errorf("could not find %q", ref)
	}
	if decl.Position().Filename == "" {
		return fmt.Errorf("could not find location of %q", ref)
	}
	if i.Editor == "auto" {
		if i.TerminalProgram == "vscode" {
			i.Editor = "vscode"
		} else if strings.Contains(i.TerminalEmulator, "JetBrains") {
			i.Editor = "intellij"
		} else {
			return fmt.Errorf(`could not auto choose default editor, use one of the following values:
	vscode
	intellij`)
		}
	}
	switch i.Editor {
	case "vscode":
		return openVisualStudioCode(ctx, decl.Position(), pc.Root())
	case "intellij":
		return openIntelliJ(ctx, decl.Position(), pc.Root())
	default:
		// TODO: print out valid editors
		return fmt.Errorf("unsupported editor %q", i.Editor)
	}
}

// mergedSchemaFromResp parses the schema from the response and applies any changesets.
// modules that are removed by a changeset stay in the schema.
func mergedSchemaFromResp(resp *ftlv1.GetSchemaResponse) (*schema.Schema, error) {
	sch, err := schema.FromProto(resp.Schema)
	if err != nil {
		return nil, fmt.Errorf("could not parse schema: %w", err)
	}
	for _, cs := range resp.Changesets {
		fmt.Printf("Considering changeset %q\n", cs.Key)

		for _, module := range cs.Modules {
			moduleSch, err := schema.ModuleFromProto(module)
			if err != nil {
				return nil, fmt.Errorf("could not parse module: %w", err)
			}

			var found bool
			for i, m := range sch.Modules {
				if m.Name != module.Name {
					continue
				}
				sch.Modules[i] = moduleSch
				found = true
				break
			}
			if !found {
				sch.Modules = append(sch.Modules, moduleSch)
			}
		}
	}
	return sch, nil
}

func openVisualStudioCode(ctx context.Context, pos schema.Position, projectRoot string) error {
	// TODO: schema filepath is has a root of `ftl/modulename`. Replace it with module root
	path := pos.Filename + ":" + fmt.Sprint(pos.Line)
	if pos.Column > 0 {
		path += ":" + fmt.Sprint(pos.Column)
	}
	err := exec.Command(ctx, log.Debug, ".", "code", projectRoot, "--goto", path).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("could not open visual studio code: %w", err)
	}
	return nil
}

func openIntelliJ(ctx context.Context, pos schema.Position, projectRoot string) error {
	// TODO: schema filepath is has a root of `ftl/modulename`. Replace it with module root
	// TODO: if idea is not available, explain how to activate it
	err := exec.Command(ctx, log.Debug, ".", "idea", projectRoot, "--line", strconv.Itoa(pos.Line), "--column", strconv.Itoa(pos.Column), pos.Filename).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("could not open IntelliJ IDEA: %w", err)
	}
	return nil
}
