package timelineclient

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/schema"
)

type Ingress struct {
	DeploymentKey   key.Deployment
	RequestKey      key.Request
	StartTime       time.Time
	Verb            *schema.Ref
	RequestMethod   string
	RequestPath     string
	RequestHeaders  http.Header
	ResponseStatus  int
	ResponseHeaders http.Header
	RequestBody     []byte
	ResponseBody    []byte
	Error           optional.Option[string]
}

var _ Event = Ingress{}

func (Ingress) clientEvent() {}
func (i Ingress) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	requestKey := i.RequestKey.String()

	requestBody := i.RequestBody
	if len(requestBody) == 0 {
		requestBody = []byte("{}")
	}

	responseBody := i.ResponseBody
	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	reqHeaderBytes, err := json.Marshal(i.RequestHeaders)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request header")
	}

	respHeaderBytes, err := json.Marshal(i.ResponseHeaders)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response header")
	}

	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_Ingress{
			Ingress: &timelinepb.IngressEvent{
				DeploymentKey:  i.DeploymentKey.String(),
				RequestKey:     &requestKey,
				Timestamp:      timestamppb.New(i.StartTime),
				VerbRef:        i.Verb.ToProto(), //nolint:forcetypeassert
				Method:         i.RequestMethod,
				Path:           i.RequestPath,
				StatusCode:     int32(i.ResponseStatus),
				Duration:       durationpb.New(time.Since(i.StartTime)),
				Request:        string(requestBody),
				RequestHeader:  string(reqHeaderBytes),
				Response:       string(responseBody),
				ResponseHeader: string(respHeaderBytes),
				Error:          i.Error.Ptr(),
			},
		},
	}, nil
}
