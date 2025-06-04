package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type schemaSaveCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dest        string   `help:"Path to schema to write (defaults to ftl-schema.json in the project root)."`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv    []string `help:"Environment variables to set for the build."`
}

func (s *schemaSaveCmd) Run(
	ctx context.Context,
	adminClient adminpbconnect.AdminServiceClient,
	schemaSource *schemaeventsource.EventSource,
	projConfig projectconfig.Config,
) error {
	logger := log.FromContext(ctx)
	if len(s.Dirs) == 0 {
		s.Dirs = projConfig.AbsModuleDirs()
	}
	if len(s.Dirs) == 0 {
		return errors.WithStack(errors.New("no directories specified"))
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "build stopped"))
	engine, err := buildengine.New(
		ctx,
		adminClient,
		schemaSource,
		projConfig,
		s.Dirs,
		false,
		buildengine.BuildEnv(s.BuildEnv),
		buildengine.Parallelism(s.Parallelism),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(engine.Modules()) == 0 {
		logger.Warnf("No modules were found to build")
		return nil
	}
	if err := engine.Build(ctx); err != nil {
		return errors.Wrap(err, "build failed")
	}
	sch, ok := engine.GetSchema()
	if !ok {
		return errors.New("schema not found")
	}
	pb := sch.ToProto()
	js, err := protojson.Marshal(pb)
	if err != nil {
		return errors.Wrap(err, "failed to JSON encode schema")
	}
	dest := s.Dest
	if dest == "" {
		dest = filepath.Join(projConfig.Root(), "ftl-schema.json")
	}
	err = os.WriteFile(dest, js, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to save schema")
	}
	logger.Infof("Wrote schema to %q", dest) //nolint
	return nil
}
