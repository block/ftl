package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/chroma/v2/quick"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/either"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mattn/go-isatty"

	"github.com/block/ftl/backend/admin"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/terminal"
	"github.com/block/ftl/internal/watch"
)

type schemaDiffCmd struct {
	OtherEndpoint url.URL `arg:"" help:"Other endpoint URL to compare against. If this is not specified then ftl will perform a diff against the local schema." optional:""`
	Color         bool    `help:"Enable colored output regardless of TTY."`
}

func (d *schemaDiffCmd) Run(
	ctx context.Context,
	currentURL *url.URL,
	projConfig projectconfig.Config,
	schemaClient admin.EnvironmentClient,
) error {
	var other *schema.Schema
	var err error
	sameModulesOnly := false
	otherEndpoint := d.OtherEndpoint.String()
	if otherEndpoint == "" {
		otherEndpoint = "Local Changes"
		sameModulesOnly = true

		other, err = localSchema(ctx, projConfig)
	} else {
		other, err = schemaForURL(ctx, schemaClient, d.OtherEndpoint)
	}
	if err != nil {
		return errors.Wrap(err, "failed to get other schema")
	}
	current, err := schemaForURL(ctx, schemaClient, *currentURL)
	if err != nil {
		return errors.Wrap(err, "failed to get current schema")
	}
	if sameModulesOnly {
		for _, realm := range current.Realms {
			tempModules := realm.Modules
			realm.Modules = []*schema.Module{}
			moduleMap := map[string]*schema.Module{}
			for _, i := range tempModules {
				moduleMap[i.Name] = i
			}
			for _, r := range other.Realms {
				if r.Name != realm.Name {
					continue
				}
				for _, i := range r.Modules {
					if mod, ok := moduleMap[i.Name]; ok {
						realm.Modules = append(realm.Modules, mod)
					}
				}
			}
		}
	}

	edits := myers.ComputeEdits(span.URIFromPath(""), current.String(), other.String())
	diff := fmt.Sprint(gotextdiff.ToUnified(currentURL.String(), otherEndpoint, current.String(), edits))

	color := d.Color || isatty.IsTerminal(os.Stdout.Fd())
	if color {
		err = quick.Highlight(os.Stdout, diff, "diff", "terminal256", "solarized-dark")
		if err != nil {
			return errors.Wrap(err, "failed to highlight diff")
		}
	} else {
		fmt.Print(diff)
	}

	// Similar to the `diff` command, exit with 1 if there are differences.
	if diff != "" {
		// Unfortunately we need to close the terminal before exit to make sure the output is printed
		// This is only applicable when we explicitly call os.Exit
		terminal.FromContext(ctx).Close()
		os.Exit(1)
	}

	return nil
}

func localSchema(ctx context.Context, projectConfig projectconfig.Config) (*schema.Schema, error) {
	errs := []error{}
	modules, err := watch.DiscoverModules(ctx, projectConfig.AbsModuleDirs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to discover module")
	}

	moduleSchemas := make(chan either.Either[*schema.Module, error], len(modules))
	defer close(moduleSchemas)

	for _, m := range modules {
		go func() {
			module, err := schema.ModuleFromProtoFile(projectConfig.SchemaPath(m.Module))
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](err)
				return
			}
			moduleSchemas <- either.LeftOf[error](module)
		}()
	}
	realm := &schema.Realm{
		Name:    "default", // TODO: projectConfig.Name,
		Modules: []*schema.Module{},
	}
	sch := &schema.Schema{Realms: []*schema.Realm{realm}}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			realm.Upsert(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))

		}
	}
	// we want schema even if there are errors as long as we have some modules
	if len(sch.InternalModules()) == 0 && len(errs) > 0 {
		return nil, errors.Wrap(errors.Join(errs...), "failed to read schema, possibly due to not building")
	}
	return sch, nil
}
func schemaForURL(ctx context.Context, schemaClient admin.EnvironmentClient, url url.URL) (*schema.Schema, error) {
	resp, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return nil, errors.Wrapf(err, "url %s: failed to get schema", url.String())
	}

	s, err := schema.FromProto(resp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrapf(err, "url %s: failed to parse schema", url.String())
	}

	return s, nil
}
