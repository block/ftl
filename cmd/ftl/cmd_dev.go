package main

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/lsp"
)

type devCmd struct {
	Parallelism    int           `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs           []string      `arg:"" help:"Base directories containing modules." type:"existingdir" optional:""`
	External       []string      `help:"Directories for libraries that require FTL module stubs." type:"existingdir" optional:""`
	Watch          time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe        bool          `help:"Do not start the FTL server." default:"false"`
	RunLsp         bool          `help:"Run the language server." default:"false"`
	ServeCmd       serveCmd      `embed:""`
	languageServer *lsp.Server
}

func (d *devCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	if len(d.Dirs) == 0 && len(d.External) == 0 {
		d.Dirs = projConfig.ModuleDirs
		d.External = projConfig.ExternalDirs
	}
	if len(d.Dirs) == 0 && len(d.External) == 0 {
		return errors.New("no directories specified")
	}

	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)

	g, ctx := errgroup.WithContext(ctx)

	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.Run(ctx)
			if err != nil {
				return err
			}
			d.ServeCmd.Stop = false
		}
		if d.ServeCmd.isRunning(ctx, client) {
			return errors.New(ftlRunningErrorMsg)
		}

		g.Go(func() error { return d.ServeCmd.Run(ctx) })
	}

	g.Go(func() error {
		err := d.ServeCmd.waitForControllerOnline(ctx, client)
		if err != nil {
			return err
		}

		var listener *buildengine.Listener
		if d.RunLsp {
			d.languageServer = lsp.NewServer(ctx)
			listener = &buildengine.Listener{
				OnBuildComplete:  d.OnBuildComplete,
				OnDeployComplete: d.OnDeployComplete,
			}
			g.Go(func() error {
				return d.languageServer.Run()
			})
		}

		engine, err := buildengine.New(ctx, client, d.Dirs, d.External, buildengine.Parallelism(d.Parallelism), buildengine.WithListener(listener))
		if err != nil {
			return err
		}
		return engine.Dev(ctx, d.Watch)
	})

	return g.Wait()
}

func (d *devCmd) OnBuildComplete(project buildengine.Project, err error) {
	d.languageServer.BuildComplete(project.Config().Dir, err)
}

func (d *devCmd) OnDeployComplete(project buildengine.Project, err error) {
	d.languageServer.DeployComplete(project.Config().Dir, err)
}
