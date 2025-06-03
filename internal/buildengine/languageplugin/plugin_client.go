package languageplugin

import (
	"context"
	"syscall"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/internal/exec"
)

type streamCancelFunc func()

type pluginClient interface {
	getDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error)

	generateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error)
	syncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error)

	build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (*connect.Response[langpb.BuildResponse], error)

	kill() error
	cmdErr() <-chan error
}

var _ pluginClient = &pluginClientImpl{}

type pluginClientImpl struct {
	plugin *plugin.Plugin[langconnect.LanguageServiceClient, ftlv1.PingRequest, ftlv1.PingResponse, *ftlv1.PingResponse]

	// channel gets closed when the plugin exits
	cmdError chan error
}

func newClientImpl(ctx context.Context, dir, language, name string) (*pluginClientImpl, error) {
	impl := &pluginClientImpl{}
	err := impl.start(ctx, dir, language, name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return impl, nil
}

// cmdPathForLanguage returns the path to the language plugin executable for the given language.
//
// This allows us to return a helpful error message if no plugin is found.
func cmdPathForLanguage(language string) (string, error) {
	cmdName := "ftl-language-" + language
	path, err := exec.LookPath(cmdName)
	if err != nil {
		return "", errors.Wrapf(err, "failed to find plugin for %s", language)
	}
	return path, nil
}

// Start launches the plugin and blocks until the plugin is ready.
func (p *pluginClientImpl) start(ctx context.Context, dir, language, name string) error {
	cmdPath, err := cmdPathForLanguage(language)
	if err != nil {
		return errors.WithStack(err)
	}
	envvars := []string{"FTL_NAME=" + name}
	plugin, cmdCtx, err := plugin.Spawn(ctx,
		log.FromContext(ctx).GetLevel(),
		name,
		name,
		dir,
		cmdPath,
		langconnect.NewLanguageServiceClient,
		true,
		plugin.WithEnvars(envvars...),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to spawn plugin for %s", name)
	}
	p.plugin = plugin

	p.cmdError = make(chan error)
	go func() {
		<-cmdCtx.Done()
		err := cmdCtx.Err()
		if err != nil {
			p.cmdError <- errors.Wrap(err, "language plugin failed")
		} else {
			p.cmdError <- errors.Errorf("language plugin ended with status 0")
		}
	}()
	return nil
}

func (p *pluginClientImpl) kill() error {
	if err := p.plugin.Cmd.Kill(syscall.SIGINT); err != nil {
		return errors.WithStack(err) //nolint:wrapcheck
	}
	return nil
}

func (p *pluginClientImpl) cmdErr() <-chan error {
	return p.cmdError
}

func (p *pluginClientImpl) getDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := p.plugin.Client.GetDependencies(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err) //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) generateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := p.plugin.Client.GenerateStubs(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err) //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) syncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := p.plugin.Client.SyncStubReferences(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err) //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (*connect.Response[langpb.BuildResponse], error) {
	if err := p.checkCmdIsAlive(); err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := p.plugin.Client.Build(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err) //nolint:wrapcheck
	}
	return resp, nil
}

func (p *pluginClientImpl) checkCmdIsAlive() error {
	select {
	case err := <-p.cmdError:
		if err == nil {
			// cmd errored with success or the channel was closed previously
			return errors.WithStack(ErrPluginNotRunning)
		}
		return errors.Join(err, ErrPluginNotRunning)
	default:
		return nil
	}
}
