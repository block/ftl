package app

import (
	"context"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/devstate"
	"github.com/block/ftl/internal/editor"
	"github.com/block/ftl/internal/projectconfig"
)

type editCmd struct {
	Ref    reflection.Ref `arg:"" help:"Language of the module to create." placeholder:"MODULE.ITEM" predictor:"decls"`
	Editor string         `arg:"" help:"Editor to open the file with." enum:"${supportedEditors}"`
}

func (editCmd) Completions() map[string][]string {
	return map[string][]string{
		"Editor": editor.SupportedEditors,
	}
}

func (i editCmd) Run(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient adminpbconnect.AdminServiceClient, pc projectconfig.Config) error {
	state, err := devstate.WaitForDevState(ctx, buildEngineClient, adminClient, false)
	if err != nil {
		return errors.Wrapf(err, "could not get dev state; is FTL running?")
	}

	odecl, omod := state.Schema.ResolveWithModule(&schema.Ref{Module: i.Ref.Module, Name: i.Ref.Name})
	decl, ok := odecl.Get()
	if !ok {
		return errors.Errorf("could not find %q", i.Ref)
	}
	if decl.Position().Filename == "" {
		return errors.Errorf("could not find file of %q", i.Ref)
	}

	mod, ok := omod.Get()
	if !ok {
		return errors.Errorf("could not find module %q", i.Ref.Module)
	}

	err = editor.OpenFileInEditor(ctx, i.Editor, decl.Position(), pc.Root(), mod)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s in %s", i.Ref, i.Editor)
	}
	return nil
}
