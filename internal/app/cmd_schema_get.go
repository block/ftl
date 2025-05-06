package app

import (
	"context"
	"fmt"
	"os"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/projectconfig"
)

type getSchemaCmd struct {
	Watch    bool     `help:"Watch for changes to the schema."`
	Protobuf bool     `help:"Output the schema as binary protobuf." xor:"format"`
	JSON     bool     `help:"Output the schema as JSON." xor:"format"`
	Modules  []string `help:"Modules to include" type:"string" optional:""`
	External bool     `help:"Outputs the externally exported part of the schema"`
}

func (g *getSchemaCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient, projConfig projectconfig.Config) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-schema-get"}))
	if err != nil {
		return errors.WithStack(err)
	}
	if g.Protobuf {
		return errors.WithStack(g.generateProto(resp))
	}
	if g.JSON {
		return errors.WithStack(g.generateJSON(resp))
	}
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	for resp.Receive() {
		msg := resp.Msg()
		notification, err := schema.NotificationFromProto(msg.Event)
		if err != nil {
			return errors.Wrap(err, "invalid notification")
		}
		sch := &schema.Schema{}
		switch e := notification.(type) {
		case *schema.FullSchemaNotification:
			sch = e.Schema
			if len(g.Modules) > 0 {
				sch, _ = sch.FilterModules(moduleNamesToRefKeys(g.Modules))
			}
			if g.External {
				sch = sch.External()
			}
			err = g.displaySchema(sch)
			if err != nil {
				return errors.WithStack(err)
			}
			if !g.Watch {
				return nil
			}
		case *schema.ChangesetCommittedNotification:
			var modules []*schema.Module
			modules = append(modules, e.Changeset.InternalRealm().Modules...)
			realm := &schema.Realm{
				Name:    projConfig.Name,
				Modules: modules,
			}
			sch.Realms = append(sch.Realms, realm)
			sch, _ = g.applyFilters(sch)
			err = g.displaySchema(sch)
			if err != nil {
				return errors.WithStack(err)
			}
		default:
			// Ignore for now
		}

	}
	if err := resp.Err(); err != nil {
		return errors.WithStack(resp.Err())
	}
	return nil
}

func (g *getSchemaCmd) displaySchema(sch *schema.Schema) error {
	for _, realm := range sch.Realms {
		fmt.Println(realm)
	}
	return nil
}

func fullSchemaFromStream(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) (*schema.Schema, error) {
	for resp.Receive() {
		msg := resp.Msg()
		notification, err := schema.NotificationFromProto(msg.Event)
		if err != nil {
			return nil, errors.Wrap(err, "invalid notification")
		}

		switch e := notification.(type) {
		case *schema.FullSchemaNotification:
			return e.Schema, nil
		default:
			// Ignore for now
		}
	}
	if err := resp.Err(); err != nil {
		return nil, errors.Wrap(err, "error receiving schema")
	}
	return nil, errors.New("no FullSchemaNotification received")
}

func (g *getSchemaCmd) generateJSON(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	sch, err := fullSchemaFromStream(resp)
	if err != nil {
		return errors.Wrap(err, "error receiving schema")
	}
	sch, missing := g.applyFilters(sch)

	data, err := protojson.Marshal(sch.ToProto())
	if err != nil {
		return errors.Wrap(err, "error marshaling schema")
	}
	fmt.Printf("%s\n", data)
	missingNames := slices.Map(missing, func(m schema.RefKey) string { return m.Module })
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return errors.Errorf("missing modules: %v", missingNames)
	}
	return nil
}

func (g *getSchemaCmd) generateProto(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	sch, err := fullSchemaFromStream(resp)
	if err != nil {
		return errors.Wrap(err, "error receiving schema")
	}
	sch, missing := g.applyFilters(sch)
	pb, err := proto.Marshal(sch.ToProto())
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = os.Stdout.Write(pb)
	if err != nil {
		return errors.WithStack(err)
	}
	missingNames := slices.Map(missing, func(m schema.RefKey) string { return m.Module })
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return errors.Errorf("missing modules: %v", missingNames)
	}
	return nil
}

func (g *getSchemaCmd) applyFilters(sch *schema.Schema) (*schema.Schema, []schema.RefKey) {
	var missing []schema.RefKey
	if len(g.Modules) > 0 {
		sch, missing = sch.FilterModules(moduleNamesToRefKeys(g.Modules))
	}
	if g.External {
		sch = sch.External()
	}
	return sch, missing
}

func moduleNamesToRefKeys(modules []string) []schema.RefKey {
	return slices.Map(modules, func(m string) schema.RefKey { return schema.RefKey{Module: m} })
}
