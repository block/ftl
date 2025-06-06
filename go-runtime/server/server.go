package server

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	leaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/go-runtime/server/query"
	"github.com/block/ftl/go-runtime/server/rpccontext"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/maps"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/rpc"
)

type UserVerbConfig struct {
	FTLRunnerEndpoint   *url.URL             `help:"FTL runner endpoint." env:"FTL_RUNNER_ENDPOINT" required:""`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	Config              []string             `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
}

// NewUserVerbServer starts a new code-generated drive for user Verbs.
//
// This function is intended to be used by the code generator.
func NewUserVerbServer(projectName string, moduleName string, handlers ...Handler) plugin.Constructor[ftlv1connect.VerbServiceHandler, UserVerbConfig] {
	return func(ctx context.Context, uc UserVerbConfig) (context.Context, ftlv1connect.VerbServiceHandler, error) {
		moduleServiceClient := rpc.Dial(ftlv1connect.NewDeploymentContextServiceClient, uc.FTLRunnerEndpoint.String(), log.Error)
		ctx = rpccontext.ContextWithClient(ctx, moduleServiceClient)
		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, uc.FTLRunnerEndpoint.String(), log.Error)
		ctx = rpccontext.ContextWithClient(ctx, verbServiceClient)
		pubClient := rpc.Dial(pubsubpbconnect.NewPublishServiceClient, uc.FTLRunnerEndpoint.String(), log.Error)
		ctx = rpccontext.ContextWithClient(ctx, pubClient)
		leaseClient := rpc.Dial(leaseconnect.NewLeaseServiceClient, uc.FTLRunnerEndpoint.String(), log.Error)
		ctx = rpccontext.ContextWithClient(ctx, leaseClient)
		queryClient := rpc.Dial(querypbconnect.NewQueryServiceClient, uc.FTLRunnerEndpoint.String(), log.Error)
		ctx = rpccontext.ContextWithClient(ctx, queryClient)

		moduleContextSupplier := deploymentcontext.NewDeploymentContextSupplier(moduleServiceClient)
		// FTL_DEPLOYMENT is set by the FTL runtime.
		dynamicCtx, err := deploymentcontext.NewDynamicContext(ctx, moduleContextSupplier, os.Getenv("FTL_DEPLOYMENT"))
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not get config")
		}

		ctx = dynamicCtx.ApplyToContext(ctx)
		ctx = internal.WithContext(ctx, internal.New(dynamicCtx))

		err = observability.Init(ctx, true, projectName, moduleName, "HEAD", uc.ObservabilityConfig)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not initialize metrics")
		}
		hmap := maps.FromSlice(handlers, func(h Handler) (reflection.Ref, Handler) { return h.ref, h })
		return ctx, &moduleServer{handlers: hmap}, nil
	}
}

// Handler for a Verb.
type Handler struct {
	ref reflection.Ref
	fn  func(ctx context.Context, req []byte, metadata map[internal.MetadataKey]string) ([]byte, error)
}

func HandleCall[Req, Resp any](module string, verb string) Handler {
	ref := reflection.Ref{Module: module, Name: verb}
	return Handler{
		ref: ref,
		fn: func(ctx context.Context, reqdata []byte, metadata map[internal.MetadataKey]string) ([]byte, error) {
			ctx = internal.ContextWithCallMetadata(ctx, metadata)

			// Decode request.
			var req Req
			err := encoding.Unmarshal(reqdata, &req)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid request to verb %s", ref)
			}
			ctx = observability.AddSpanContextToLogger(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "failed to add workload identity to context")
			}

			// InvokeVerb Verb.
			resp, err := InvokeVerb[Req, Resp](ref)(ctx, req)
			if err != nil {
				return nil, errors.Wrapf(err, "call to verb %s failed", ref)
			}

			respdata, err := encoding.Marshal(resp)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return respdata, nil
		},
	}
}
func HandleSink[Req any](module string, verb string) Handler {
	return HandleCall[Req, ftl.Unit](module, verb)
}

func HandleSource[Resp any](module string, verb string) Handler {
	return HandleCall[ftl.Unit, Resp](module, verb)
}

func HandleEmpty(module string, verb string) Handler {
	return HandleCall[ftl.Unit, ftl.Unit](module, verb)
}

func InvokeVerb[Req, Resp any](ref reflection.Ref) func(ctx context.Context, req Req) (resp Resp, err error) {
	return func(ctx context.Context, req Req) (resp Resp, err error) {
		request := optional.Some[any](req)
		if reflect.TypeFor[Req]() == reflect.TypeFor[ftl.Unit]() {
			request = optional.None[any]()
		}

		out, err := reflection.CallVerb(reflection.Ref{Module: ref.Module, Name: ref.Name})(ctx, request)
		if err != nil {
			return resp, errors.WithStack(err)
		}

		var respValue any
		if r, ok := out.Get(); ok {
			respValue = r
		} else {
			respValue = ftl.Unit{}
		}
		resp, ok := respValue.(Resp)
		if !ok {
			return resp, errors.Errorf("unexpected response type from verb %s: %T, expected %T", ref, resp, reflect.New(reflect.TypeFor[Resp]()).Interface())
		}
		return resp, errors.WithStack(err)
	}
}

func VerbClient[Verb, Req, Resp any]() reflection.VerbResource {
	fnCall := call[Verb, Req, Resp]()
	return func() reflect.Value {
		return reflect.ValueOf(fnCall)
	}
}

func SinkClient[Verb, Req any]() reflection.VerbResource {
	fnCall := call[Verb, Req, ftl.Unit]()
	sink := func(ctx context.Context, req Req) error {
		_, err := fnCall(ctx, req)
		return errors.WithStack(err)
	}
	return func() reflect.Value {
		return reflect.ValueOf(sink)
	}
}

func SourceClient[Verb, Resp any]() reflection.VerbResource {
	fnCall := call[Verb, ftl.Unit, Resp]()
	source := func(ctx context.Context) (Resp, error) {
		return errors.WithStack2(fnCall(ctx, ftl.Unit{}))
	}
	return func() reflect.Value {
		return reflect.ValueOf(source)
	}
}

func EmptyClient[Verb any]() reflection.VerbResource {
	fnCall := call[Verb, ftl.Unit, ftl.Unit]()
	source := func(ctx context.Context) error {
		_, err := fnCall(ctx, ftl.Unit{})
		return errors.WithStack(err)
	}
	return func() reflect.Value {
		return reflect.ValueOf(source)
	}
}

func call[Verb, Req, Resp any]() func(ctx context.Context, req Req) (resp Resp, err error) {
	typ := reflect.TypeFor[Verb]()
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("Cannot register %s: expected function, got %s", typ, typ.Kind()))
	}
	callee := reflection.TypeRef[Verb]()
	callee.Name = strings.TrimSuffix(callee.Name, "Client")
	return func(ctx context.Context, req Req) (resp Resp, err error) {
		ref := reflection.Ref{Module: callee.Module, Name: callee.Name}
		moduleCtx := deploymentcontext.FromContext(ctx).CurrentContext()
		override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: ref.Module, Name: ref.Name})
		if err != nil {
			return resp, errors.Wrapf(err, "%s", ref)
		}
		if behavior, ok := override.Get(); ok {
			uncheckedResp, err := behavior.Call(ctx, deploymentcontext.Verb(widenVerb(InvokeVerb[Req, Resp](ref))), req)
			if err != nil {
				return resp, errors.Wrapf(err, "%s", ref)
			}
			if r, ok := uncheckedResp.(Resp); ok {
				return r, nil
			}
			return resp, errors.Errorf("%s: overridden verb had invalid response type %T, expected %v", ref,
				uncheckedResp, reflect.TypeFor[Resp]())
		}

		reqData, err := encoding.Marshal(req)
		if err != nil {
			return resp, errors.Wrapf(err, "%s: failed to marshal request", callee)
		}

		client := rpccontext.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
		callReq := &ftlv1.CallRequest{Verb: callee.ToProto(), Body: reqData}
		// propagate transaction ID to the next call
		if callMd, ok := internal.MaybeCallMetadataFromContext(ctx).Get(); ok {
			if txnID, ok := callMd[query.TransactionMetadataKey]; ok {
				callReq.Metadata = &ftlv1.Metadata{Values: []*ftlv1.Metadata_Pair{query.GetTransactionMetadata(txnID)}}
			}
		}
		cresp, err := client.Call(ctx, connect.NewRequest(callReq))
		if err != nil {
			return resp, errors.Wrapf(err, "%s: failed to call Verb", callee)
		}
		switch cresp := cresp.Msg.Response.(type) {
		case *ftlv1.CallResponse_Error_:
			return resp, errors.Errorf("%s: %s", callee, cresp.Error.Message)

		case *ftlv1.CallResponse_Body:
			err = encoding.Unmarshal(cresp.Body, &resp)
			if err != nil {
				return resp, errors.Wrapf(err, "%s: failed to decode response", callee)
			}
			return resp, nil

		default:
			panic(fmt.Sprintf("%s: invalid response type %T", callee, cresp))
		}
	}
}

var _ ftlv1connect.VerbServiceHandler = (*moduleServer)(nil)

// This is the server that is compiled into the same binary as user-defined Verbs.
type moduleServer struct {
	handlers map[reflection.Ref]Handler
}

func (m *moduleServer) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (response *connect.Response[ftlv1.CallResponse], err error) {
	logger := log.FromContext(ctx)
	// Recover from panics and return an error ftlv1.CallResponse.
	defer func() {
		if r := recover(); r != nil {
			var err error
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = errors.Errorf("%v", r)
			}
			stack := string(debug.Stack())
			logger.Errorf(err, "panic in verb %s.%s", req.Msg.Verb.Module, req.Msg.Verb.Name)
			response = connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{
				Message: err.Error(),
				Stack:   &stack,
			}}})
		}
	}()

	handler, ok := m.handlers[reflection.RefFromProto(req.Msg.Verb)]
	if !ok {
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("verb %s.%s not found", req.Msg.Verb.Module, req.Msg.Verb.Name)))
	}

	metadata := map[internal.MetadataKey]string{}
	if req.Msg.Metadata != nil {
		for _, pair := range req.Msg.Metadata.Values {
			metadata[internal.MetadataKey(pair.Key)] = pair.Value
		}
	}

	ctx, err = addWorkloadIdentity(ctx, req.Header())

	respdata, err := handler.fn(ctx, req.Msg.Body, metadata)
	if err != nil {
		// This makes me slightly ill.
		return connect.NewResponse(&ftlv1.CallResponse{
			Response: &ftlv1.CallResponse_Error_{Error: &ftlv1.CallResponse_Error{Message: err.Error()}},
		}), nil
	}

	return connect.NewResponse(&ftlv1.CallResponse{
		Response: &ftlv1.CallResponse_Body{Body: respdata},
	}), nil
}

func (m *moduleServer) Ping(_ context.Context, _ *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func widenVerb[Req, Resp any](verb ftl.Verb[Req, Resp]) ftl.Verb[any, any] {
	return func(ctx context.Context, uncheckedReq any) (any, error) {
		req, ok := uncheckedReq.(Req)
		if !ok {
			return nil, errors.Errorf("invalid request type %T for %v, expected %v", uncheckedReq, reflection.FuncRef(verb), reflect.TypeFor[Req]())
		}
		return errors.WithStack2(verb(ctx, req))
	}
}
