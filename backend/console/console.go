package console

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	ftlversion "github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/goose"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	consolepb "github.com/block/ftl/backend/protos/xyz/block/ftl/console/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/console/v1/consolepbconnect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	frontend "github.com/block/ftl/frontend/console"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/editor"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type Config struct {
	ContentTime time.Time `help:"Time to use for console resource timestamps." default:"${timestamp=1970-01-01T00:00:00Z}"`
}

type Service struct {
	schemaEventSource *schemaeventsource.EventSource
	timelineClient    *timelineclient.Client
	adminClient       admin.EnvironmentClient
	callClient        routing.CallClient
	buildEngineClient buildenginepbconnect.BuildEngineServiceClient
	bind              *url.URL
	config            Config
	pc                optional.Option[projectconfig.Config]
	isLocalDev        bool
}

var _ consolepbconnect.ConsoleServiceHandler = (*Service)(nil)

func New(
	eventSource *schemaeventsource.EventSource,
	timelineClient *timelineclient.Client,
	adminClient admin.EnvironmentClient,
	client routing.CallClient,
	buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	bind *url.URL,
	config Config,
	pc optional.Option[projectconfig.Config],
	isLocalDev bool,
) *Service {

	return &Service{
		schemaEventSource: eventSource,
		timelineClient:    timelineClient,
		adminClient:       adminClient,
		callClient:        client,
		buildEngineClient: buildEngineClient,
		bind:              bind,
		config:            config,
		pc:                pc,
		isLocalDev:        isLocalDev,
	}
}

func (s *Service) StartServices(ctx context.Context) ([]rpc.Option, error) {
	logger := log.FromContext(ctx).Scope("console")
	ctx = log.ContextWithLogger(ctx, logger)

	consoleHandler, err := frontend.Server(ctx, s.config.ContentTime, s.bind)
	if err != nil {
		return nil, errors.Wrap(err, "could not start console")
	}
	logger.Infof("Web console available at: %s", s.bind) //nolint

	return []rpc.Option{
		rpc.GRPC(consolepbconnect.NewConsoleServiceHandler, s),
		rpc.HTTP("/", consoleHandler),
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func visitNode(sch *schema.Schema, n schema.Node, verbString *string) error {
	return errors.WithStack(schema.Visit(n, func(n schema.Node, next func() error) error {
		switch n := n.(type) {
		case *schema.Ref:
			if decl, ok := sch.Resolve(n).Get(); ok {
				*verbString += decl.String() + "\n\n"
				err := visitNode(sch, decl, verbString)
				if err != nil {
					return errors.WithStack(err)
				}
			}

		default:
		}
		return errors.WithStack(next())
	}))
}

func verbSchemaString(sch *schema.Schema, verb *schema.Verb) (string, error) {
	var verbString string
	err := visitNode(sch, verb.Request, &verbString)
	if err != nil {
		return "", errors.WithStack(err)
	}
	// Don't print the response if it's the same as the request.
	if !verb.Response.Equal(verb.Request) {
		err = visitNode(sch, verb.Response, &verbString)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}
	verbString += verb.String()
	return verbString, nil
}

func (s *Service) GetModules(ctx context.Context, req *connect.Request[consolepb.GetModulesRequest]) (*connect.Response[consolepb.GetModulesResponse], error) {
	sch := s.schemaEventSource.CanonicalView()

	allowed := map[string]bool{}
	var modules []*consolepb.Module
	for _, mod := range sch.InternalModules() {
		if mod.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
			continue
		}
		allowed[mod.Name] = true

		pbModuleFields, err := moduleFromDecls(mod.Decls, sch, mod)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to process declarations for module %s", mod.Name)
		}

		modules = append(modules, &consolepb.Module{
			Name:        mod.Name,
			Module:      mod.ToProto(),
			Runtime:     mod.Runtime.ToProto(),
			Schema:      mod.String(),
			Verbs:       pbModuleFields.Verbs,
			Data:        pbModuleFields.Data,
			Secrets:     pbModuleFields.Secrets,
			Configs:     pbModuleFields.Configs,
			Databases:   pbModuleFields.Databases,
			Enums:       pbModuleFields.Enums,
			Topics:      pbModuleFields.Topics,
			Typealiases: pbModuleFields.Typealiases,
		})
	}

	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return nil, errors.Wrap(err, "failed to sort modules")
	}
	topology := &consolepb.Topology{
		Levels: make([]*consolepb.TopologyGroup, len(sorted)),
	}
	for i, level := range sorted {
		gLevels := []string{}
		for _, i := range level {
			if allowed[i] {
				gLevels = append(gLevels, i)
			}
		}
		group := &consolepb.TopologyGroup{
			Modules: gLevels,
		}
		topology.Levels[i] = group
	}

	return connect.NewResponse(&consolepb.GetModulesResponse{
		Modules:  modules,
		Topology: topology,
	}), nil
}

func toEdges(sch *schema.Schema, module string, name string) *consolepb.Edges {
	graph := schema.Graph(sch)
	key := schema.RefKey{
		Module: module,
		Name:   name,
	}

	if node, ok := graph[key]; ok {
		return &consolepb.Edges{
			In:  refsToProto(node.In),
			Out: refsToProto(node.Out),
		}
	}
	return &consolepb.Edges{}
}

func refsToProto(refs []schema.RefKey) []*schemapb.Ref {
	out := make([]*schemapb.Ref, len(refs))
	for i, ref := range refs {
		out[i] = ref.ToProto()
	}
	return out
}

func configFromDecl(decl *schema.Config, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Config {
	return &consolepb.Config{
		Config: decl.ToProto(),
		Edges:  toEdges(sch, module, decl.Name),
		Schema: decl.String(),
		Git:    git.ToProto(),
	}
}

func dataFromDecl(decl *schema.Data, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Data {
	return &consolepb.Data{
		Data:   decl.ToProto(),
		Schema: decl.String(),
		Edges:  toEdges(sch, module, decl.Name),
		Git:    git.ToProto(),
	}
}

func databaseFromDecl(decl *schema.Database, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Database {
	return &consolepb.Database{
		Database: decl.ToProto(),
		Edges:    toEdges(sch, module, decl.Name),
		Schema:   decl.String(),
		Git:      git.ToProto(),
	}
}

func enumFromDecl(decl *schema.Enum, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Enum {
	return &consolepb.Enum{
		Enum:   decl.ToProto(),
		Edges:  toEdges(sch, module, decl.Name),
		Schema: decl.String(),
		Git:    git.ToProto(),
	}
}

func topicFromDecl(decl *schema.Topic, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Topic {
	return &consolepb.Topic{
		Topic:  decl.ToProto(),
		Edges:  toEdges(sch, module, decl.Name),
		Schema: decl.String(),
		Git:    git.ToProto(),
	}
}

func typealiasFromDecl(decl *schema.TypeAlias, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.TypeAlias {
	return &consolepb.TypeAlias{
		Typealias: decl.ToProto(),
		Edges:     toEdges(sch, module, decl.Name),
		Schema:    decl.String(),
		Git:       git.ToProto(),
	}
}

func secretFromDecl(decl *schema.Secret, sch *schema.Schema, module string, git *schema.MetadataGit) *consolepb.Secret {
	return &consolepb.Secret{
		Secret: decl.ToProto(),
		Edges:  toEdges(sch, module, decl.Name),
		Schema: decl.String(),
		Git:    git.ToProto(),
	}
}

func verbFromDecl(decl *schema.Verb, sch *schema.Schema, module string, git *schema.MetadataGit) (*consolepb.Verb, error) {
	v := decl.ToProto()
	var jsonRequestSchema string
	if decl.Request != nil {
		if requestData, ok := decl.Request.(*schema.Ref); ok {
			jsonSchema, err := schema.RequestResponseToJSONSchema(sch, *requestData)
			if err != nil {
				return nil, errors.Wrap(err, "failed to retrieve JSON schema")
			}
			jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
			if err != nil {
				return nil, errors.Wrap(err, "failed to indent JSON schema")
			}
			jsonRequestSchema = string(jsonData)
		}
	}

	schemaString, err := verbSchemaString(sch, decl)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &consolepb.Verb{
		Verb:              v,
		Schema:            schemaString,
		JsonRequestSchema: jsonRequestSchema,
		Edges:             toEdges(sch, module, decl.Name),
		Git:               git.ToProto(),
	}, nil
}

func moduleFromDeployment(deployment *schema.Module, sch *schema.Schema) (*consolepb.Module, error) {
	module, err := moduleFromDecls(deployment.Decls, sch, deployment)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	module.Name = deployment.Name
	module.Schema = deployment.String()
	module.Runtime = deployment.Runtime.ToProto()
	return module, nil
}

func moduleFromDecls(decls []schema.Decl, sch *schema.Schema, module *schema.Module) (*consolepb.Module, error) {
	var configs []*consolepb.Config
	var data []*consolepb.Data
	var databases []*consolepb.Database
	var enums []*consolepb.Enum
	var topics []*consolepb.Topic
	var typealiases []*consolepb.TypeAlias
	var secrets []*consolepb.Secret
	var verbs []*consolepb.Verb

	var git *schema.MetadataGit
	for _, metaItem := range module.Metadata {
		if concreteGitMeta, ok := metaItem.(*schema.MetadataGit); ok {
			git = concreteGitMeta
			break
		}
	}

	for _, d := range decls {
		switch decl := d.(type) {
		case *schema.Config:
			config := configFromDecl(decl, sch, module.Name, git)
			configs = append(configs, config)

		case *schema.Data:
			data = append(data, dataFromDecl(decl, sch, module.Name, git))

		case *schema.Database:
			database := databaseFromDecl(decl, sch, module.Name, git)
			databases = append(databases, database)

		case *schema.Enum:
			enum := enumFromDecl(decl, sch, module.Name, git)
			enums = append(enums, enum)

		case *schema.Topic:
			topic := topicFromDecl(decl, sch, module.Name, git)
			topics = append(topics, topic)

		case *schema.Secret:
			secret := secretFromDecl(decl, sch, module.Name, git)
			secrets = append(secrets, secret)

		case *schema.TypeAlias:
			typealias := typealiasFromDecl(decl, sch, module.Name, git)
			typealiases = append(typealiases, typealias)

		case *schema.Verb:
			verb, err := verbFromDecl(decl, sch, module.Name, git)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			verbs = append(verbs, verb)
		}
	}

	return &consolepb.Module{
		Configs:     configs,
		Data:        data,
		Databases:   databases,
		Enums:       enums,
		Topics:      topics,
		Typealiases: typealiases,
		Secrets:     secrets,
		Verbs:       verbs,
		Git:         git.ToProto(),
	}, nil
}

func (s *Service) StreamModules(ctx context.Context, req *connect.Request[consolepb.StreamModulesRequest], stream *connect.ServerStream[consolepb.StreamModulesResponse]) error {
	err := s.sendStreamModulesResp(stream)
	if err != nil {
		return errors.WithStack(err)
	}

	events := s.schemaEventSource.Subscribe(ctx)
	for range channels.IterContext(ctx, events) {
		err = s.sendStreamModulesResp(stream)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// filterDeployments removes any duplicate modules by selecting the deployment with the
// latest CreatedAt.
func (s *Service) filterDeployments(unfilteredDeployments *schema.Realm) []*schema.Module {
	latest := make(map[string]*schema.Module)

	for _, deployment := range unfilteredDeployments.Modules {
		if deployment.Runtime == nil || deployment.Runtime.Deployment == nil || deployment.Runtime.Deployment.GetDeploymentKey().IsZero() {
			continue
		}
		if existing, found := latest[deployment.Name]; !found || deployment.Runtime.Base.CreateTime.After(existing.Runtime.Base.CreateTime) {
			latest[deployment.Name] = deployment
		}
	}

	var result []*schema.Module
	for _, value := range latest {
		result = append(result, value)
	}

	return result
}

func (s *Service) sendStreamModulesResp(stream *connect.ServerStream[consolepb.StreamModulesResponse]) error {
	unfilteredSchema := s.schemaEventSource.CanonicalView()

	realms := []*schema.Realm{}
	for _, realm := range unfilteredSchema.Realms {
		realms = append(realms, &schema.Realm{
			External: realm.External,
			Name:     realm.Name,
			Modules:  s.filterDeployments(realm),
		})
	}

	sch := &schema.Schema{Realms: realms}
	builtin := schema.Builtins()
	for _, realm := range sch.InternalRealms() {
		realm.Modules = append(realm.Modules, builtin)
	}

	// Get topology
	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return errors.Wrap(err, "failed to sort modules")
	}
	topology := &consolepb.Topology{
		Levels: make([]*consolepb.TopologyGroup, len(sorted)),
	}
	for i, level := range sorted {
		group := &consolepb.TopologyGroup{
			Modules: level,
		}
		topology.Levels[i] = group
	}

	var modules []*consolepb.Module
	for _, deployment := range sch.InternalModules() {
		if deployment.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
			continue
		}
		module, err := moduleFromDeployment(deployment, sch)
		if err != nil {
			return errors.WithStack(err)
		}
		modules = append(modules, module)
	}

	builtinModule, err := moduleFromDecls(builtin.Decls, sch, builtin)
	if err != nil {
		return errors.WithStack(err)
	}
	builtinModule.Name = builtin.Name
	builtinModule.Schema = builtin.String()
	builtinModule.Runtime = builtin.Runtime.ToProto()
	modules = append(modules, builtinModule)

	err = stream.Send(&consolepb.StreamModulesResponse{
		Modules:  modules,
		Topology: topology,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send StreamModulesResponse to stream")
	}

	return nil
}

func graph(sch *schema.Schema) map[string][]string {
	out := make(map[string][]string)
	for _, module := range sch.InternalModules() {
		buildGraph(sch, module, out)
	}
	return out
}

// buildGraph recursively builds the dependency graph
func buildGraph(sch *schema.Schema, module *schema.Module, out map[string][]string) {
	out[module.Name] = module.Imports()
	for _, dep := range module.Imports() {
		var depModule *schema.Module
		for _, m := range sch.InternalModules() {
			if m.String() == dep {
				depModule = m
				break
			}
		}
		if depModule != nil {
			buildGraph(sch, module, out)
		}
	}
}

func (s *Service) GetConfig(ctx context.Context, req *connect.Request[consolepb.GetConfigRequest]) (*connect.Response[consolepb.GetConfigResponse], error) {
	resp, err := s.adminClient.ConfigGet(ctx, connect.NewRequest(&adminpb.ConfigGetRequest{
		Ref: &adminpb.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
	}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	return connect.NewResponse(&consolepb.GetConfigResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (s *Service) SetConfig(ctx context.Context, req *connect.Request[consolepb.SetConfigRequest]) (*connect.Response[consolepb.SetConfigResponse], error) {
	_, err := s.adminClient.ConfigSet(ctx, connect.NewRequest(&adminpb.ConfigSetRequest{
		Ref: &adminpb.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to set config")
	}
	return connect.NewResponse(&consolepb.SetConfigResponse{}), nil
}

func (s *Service) GetSecret(ctx context.Context, req *connect.Request[consolepb.GetSecretRequest]) (*connect.Response[consolepb.GetSecretResponse], error) {
	resp, err := s.adminClient.SecretGet(ctx, connect.NewRequest(&adminpb.SecretGetRequest{
		Ref: &adminpb.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
	}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret")
	}
	return connect.NewResponse(&consolepb.GetSecretResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (s *Service) SetSecret(ctx context.Context, req *connect.Request[consolepb.SetSecretRequest]) (*connect.Response[consolepb.SetSecretResponse], error) {
	_, err := s.adminClient.SecretSet(ctx, connect.NewRequest(&adminpb.SecretSetRequest{
		Ref: &adminpb.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to set secret")
	}

	return connect.NewResponse(&consolepb.SetSecretResponse{}), nil
}

func (s *Service) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	resp, err := s.timelineClient.GetTimeline(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timeline from service")
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) StreamTimeline(ctx context.Context, req *connect.Request[timelinepb.StreamTimelineRequest], out *connect.ServerStream[timelinepb.StreamTimelineResponse]) error {
	stream, err := s.timelineClient.StreamTimeline(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return errors.Wrap(err, "failed to stream timeline from service")
	}
	defer stream.Close()
	for stream.Receive() {
		msg := stream.Msg()
		err = out.Send(msg)
		if err != nil {
			return errors.Wrap(err, "failed to send message")
		}
	}
	if stream.Err() != nil {
		return errors.Wrap(stream.Err(), "error streaming timeline from service")
	}
	return nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	ref, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse verb")
	}

	if err := schema.ValidateJSONCall(req.Msg.Body, ref, s.schemaEventSource.CanonicalView()); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	resp, err := s.callClient.Call(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, errors.Wrap(err, "failed to call verb")
	}
	return connect.NewResponse(resp.Msg), nil
}

// StreamEngineEvents implements consolepbconnect.ConsoleServiceHandler.
func (s *Service) StreamEngineEvents(ctx context.Context, req *connect.Request[buildenginepb.StreamEngineEventsRequest], stream *connect.ServerStream[buildenginepb.StreamEngineEventsResponse]) error {
	engineEvents, err := s.buildEngineClient.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{
		ReplayHistory: req.Msg.ReplayHistory,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to start build events stream")
	}

	for engineEvents.Receive() {
		msg := engineEvents.Msg()
		err = stream.Send(msg)
		if err != nil {
			return errors.Wrap(err, "failed to send message")
		}
	}

	if err := engineEvents.Err(); err != nil {
		return errors.Wrap(err, "error streaming build events")
	}
	return nil
}

func (s *Service) GetInfo(ctx context.Context, _ *connect.Request[consolepb.GetInfoRequest]) (*connect.Response[consolepb.GetInfoResponse], error) {
	return connect.NewResponse(&consolepb.GetInfoResponse{
		Version:    ftlversion.Version,
		BuildTime:  ftlversion.Timestamp.Format(time.RFC3339),
		IsLocalDev: s.isLocalDev,
	}), nil
}

func (s *Service) ExecuteGoose(ctx context.Context, req *connect.Request[consolepb.ExecuteGooseRequest], stream *connect.ServerStream[consolepb.ExecuteGooseResponse]) error {
	if req.Msg.Prompt == "" {
		return errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.New("prompt cannot be empty")))
	}

	logger := log.FromContext(ctx).Scope("console")
	client := goose.NewClient()

	err := client.Execute(ctx, req.Msg.Prompt, func(msg goose.Message) {
		var source consolepb.ExecuteGooseResponse_Source
		switch msg.Source {
		case goose.SourceStdout:
			source = consolepb.ExecuteGooseResponse_SOURCE_STDOUT
		case goose.SourceStderr:
			source = consolepb.ExecuteGooseResponse_SOURCE_STDERR
		case goose.SourceCompletion:
			source = consolepb.ExecuteGooseResponse_SOURCE_COMPLETION
		}

		err := stream.Send(&consolepb.ExecuteGooseResponse{
			Response: msg.Content,
			Source:   source,
		})
		if err != nil {
			logger.Debugf("failed to send response: %v", err)
		}
	})
	if err != nil {
		return errors.Wrap(err, "failed to execute goose")
	}
	return nil
}

// OpenFileInEditor opens the specified file path at the given position in the selected editor.
func (s *Service) OpenFileInEditor(ctx context.Context, req *connect.Request[consolepb.OpenFileInEditorRequest]) (*connect.Response[consolepb.OpenFileInEditorResponse], error) {
	pc, ok := s.pc.Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("opening files in editor requires a project context (ftl-project.toml)"))
	}

	msg := req.Msg
	pos := schema.Position{
		Filename: msg.Path,
		Line:     int(msg.Line),
		Column:   int(msg.Column),
	}

	err := editor.OpenFileInEditor(ctx, msg.Editor, pos, pc.Root(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %q in %s", pos.Filename, msg.Editor)
	}
	return connect.NewResponse(&consolepb.OpenFileInEditorResponse{}), nil
}
