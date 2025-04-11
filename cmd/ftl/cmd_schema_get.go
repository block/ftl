package main

import (
	"context"
	"fmt"
	"os"
	"slices"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
)

type getSchemaCmd struct {
	Watch    bool     `help:"Watch for changes to the schema."`
	Protobuf bool     `help:"Output the schema as binary protobuf." xor:"format"`
	JSON     bool     `help:"Output the schema as JSON." xor:"format"`
	Modules  []string `help:"Modules to include" type:"string" optional:""`
}

func (g *getSchemaCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-schema-get"}))
	if err != nil {
		return err
	}
	if g.Protobuf {
		return g.generateProto(resp)
	}
	if g.JSON {
		return g.generateJSON(resp)
	}
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			sch, err := schema.SchemaFromProto(e.FullSchemaNotification.Schema)
			if err != nil {
				return fmt.Errorf("invalid schema: %w", err)
			}
			err = g.handleSchema(sch)
			if err != nil {
				return err
			}
			if !g.Watch {
				return nil
			}
		case *schemapb.Notification_ChangesetCommittedNotification:
			var modules []*schema.Module
			for _, module := range e.ChangesetCommittedNotification.Changeset.RealmChanges[0].Modules {
				m, err := schema.ModuleFromProto(module)
				if err != nil {
					return fmt.Errorf("invalid module: %w", err)
				}
				modules = append(modules, m)
			}
			realm := &schema.Realm{
				Name:     "default", // TODO: implement
				External: false,
				Modules:  modules,
			}
			err = g.handleSchema(&schema.Schema{Realms: []*schema.Realm{realm}})
			if err != nil {
				return err
			}
		default:
			// Ignore for now
		}

	}
	if err := resp.Err(); err != nil {
		return resp.Err()
	}
	return nil
}

func (g *getSchemaCmd) handleSchema(sch *schema.Schema) error {
	for _, realm := range sch.Realms {
		fmt.Println(realm)
	}
	return nil
}

func (g *getSchemaCmd) generateJSON(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}

	schema := &schemapb.Schema{}
msgloop:
	for resp.Receive() {
		msg := resp.Msg()

		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, realm := range e.FullSchemaNotification.Schema.Realms {
				protoRealm := &schemapb.Realm{
					Name:     realm.Name,
					External: realm.External,
				}
				for _, module := range realm.Modules {
					if len(g.Modules) == 0 || slices.Contains(g.Modules, module.Name) {
						protoRealm.Modules = append(protoRealm.Modules, module)
						delete(remainingNames, module.Name)
					}
				}
				if len(protoRealm.Modules) > 0 {
					schema.Realms = append(schema.Realms, protoRealm)
				}
			}
			break msgloop
		default:
			// Ignore for now
		}
	}
	if err := resp.Err(); err != nil {
		return fmt.Errorf("error receiving schema: %w", err)
	}
	data, err := protojson.Marshal(schema)
	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}
	fmt.Printf("%s\n", data)
	missingNames := maps.Keys(remainingNames)
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return fmt.Errorf("missing modules: %v", missingNames)
	}
	return nil
}

func (g *getSchemaCmd) generateProto(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	schema := &schemapb.Schema{Realms: []*schemapb.Realm{{}}}
	for resp.Receive() {
		msg := resp.Msg()

		switch e := msg.Event.Value.(type) {
		case *schemapb.Notification_FullSchemaNotification:
			for _, realm := range e.FullSchemaNotification.Schema.Realms {
				protoRealm := &schemapb.Realm{
					Name:     realm.Name,
					External: realm.External,
				}
				for _, module := range realm.Modules {
					if len(g.Modules) == 0 || slices.Contains(g.Modules, module.Name) {
						protoRealm.Modules = append(protoRealm.Modules, module)
						delete(remainingNames, module.Name)
					}
				}
				if len(protoRealm.Modules) > 0 {
					schema.Realms = append(schema.Realms, protoRealm)
				}
			}
		default:
			// Ignore for now
		}
	}
	if err := resp.Err(); err != nil {
		return err
	}
	pb, err := proto.Marshal(schema)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(pb)
	if err != nil {
		return err
	}
	missingNames := maps.Keys(remainingNames)
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return fmt.Errorf("missing modules: %v", missingNames)
	}
	return nil
}
