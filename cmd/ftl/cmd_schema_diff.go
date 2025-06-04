package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/errors"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mattn/go-isatty"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/terminal"
)

type schemaDiffCmd struct {
	From  string `arg:"" optional:"" help:"Schema source to diff from, either an FTL endpoint URL or a file created by 'ftl schema save' (defaults to 'ftl-schema.json')."`
	To    string `arg:"" optional:"" help:"Schema source representing the current schema, either an FTL endpoint URL or a file created by 'ftl schema save' (defaults to running cluster)."`
	Color bool   `help:"Enable colored output regardless of TTY."`
}

func (d *schemaDiffCmd) Run(
	ctx context.Context,
	endpoint *url.URL,
	projConfig projectconfig.Config,
) error {
	from := d.From
	if from == "" {
		from = filepath.Join(projConfig.Root(), "ftl-schema.json")
	}
	to := d.To
	if to == "" {
		to = endpoint.String()
	}

	current, err := retrieveSchema(ctx, from)
	if err != nil {
		return errors.Wrap(err, "failed to get current schema")
	}
	other, err := retrieveSchema(ctx, to)
	if err != nil {
		return errors.Wrap(err, "failed to get other schema")
	}
	edits := myers.ComputeEdits(span.URIFromPath(""), current.String(), other.String())
	diff := fmt.Sprint(gotextdiff.ToUnified(from, to, current.String(), edits))

	if diff == "" {
		return nil
	}

	color := d.Color || isatty.IsTerminal(os.Stdout.Fd())
	if color {
		err = quick.Highlight(os.Stdout, diff, "diff", "terminal256", "solarized-dark")
		if err != nil {
			return errors.Wrap(err, "failed to highlight diff")
		}
	} else {
		fmt.Print(diff)
	}

	// Similar to the `diff` command, exit with 2 if there are differences.
	// Unfortunately we need to close the terminal before exit to make sure the output is printed
	// This is only applicable when we explicitly call os.Exit
	terminal.FromContext(ctx).Close()
	os.Exit(2)

	return nil
}

func retrieveSchema(ctx context.Context, source string) (*schema.Schema, error) {
	if strings.HasPrefix(source, "http") || strings.HasPrefix(source, "https") {
		return schemaFromServer(ctx, source)
	}
	return schemaFromDisk(source)
}

func schemaFromDisk(path string) (*schema.Schema, error) {
	pb, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load schema")
	}
	schemaProto := &schemapb.Schema{}
	err = protojson.Unmarshal(pb, schemaProto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schema")
	}
	sch, err := schema.FromProto(schemaProto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema")
	}
	return sch, nil
}

func schemaFromServer(ctx context.Context, url string) (*schema.Schema, error) {
	schemaClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, url, log.Warn)
	resp, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schema")
	}

	s, err := schema.FromProto(resp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema")
	}

	return s, nil
}
