package lease

import (
	"context"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	ftllease "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	leaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/rpc"
)

var _ rpc.Service = (*Service)(nil)

type Service struct {
	lock   sync.Mutex
	leases map[string]*time.Time
}

func New(ctx context.Context) *Service {

	logger := log.FromContext(ctx).Scope("lease")
	svc := &Service{
		leases: make(map[string]*time.Time),
	}
	ctx = log.ContextWithLogger(ctx, logger)
	return svc
}

func (s *Service) StartServices(ctx context.Context) ([]rpc.Option, error) {
	return []rpc.Option{rpc.GRPC(leaseconnect.NewLeaseServiceHandler, s)}, nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) AcquireLease(ctx context.Context, stream *connect.BidiStream[ftllease.AcquireLeaseRequest, ftllease.AcquireLeaseResponse]) error {
	logger := log.FromContext(ctx)
	logger.Debugf("AcquireLease called")
	c := &leaseClient{
		leases:  make(map[string]*time.Time),
		service: s,
	}
	defer c.clearLeases()
	for {
		msg, err := stream.Receive()
		if err != nil {
			logger.Errorf(err, "Could not receive lease request")
			return errors.Wrap(err, "could not receive lease request")
		}
		if len(msg.Key) == 0 {
			return errors.WithStack(errors.New("lease key must not be empty"))
		}
		logger.Debugf("Acquiring lease for: %v", msg.Key)
		success := c.handleMessage(msg.Key, msg.Ttl.AsDuration())

		if !success {
			return errors.WithStack(connect.NewError(connect.CodeResourceExhausted, errors.Errorf("lease already held")))
		}
		if err = stream.Send(&ftllease.AcquireLeaseResponse{}); err != nil {
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "could not send lease response")))
		}
	}
}

// clearLeases removes leases that were owned by a given connection
// they are only removed if the expiration time matches the one in the map
// this is to handle the case of another connection acquiring the lease after it expires
func (c *leaseClient) clearLeases() {
	s := c.service
	s.lock.Lock()
	defer s.lock.Unlock()
	for key, exp := range c.leases {
		if s.leases[key] != nil && *s.leases[key] == *exp {
			delete(s.leases, key)
		}
	}
}

func (c *leaseClient) handleMessage(keys []string, ttl time.Duration) bool {
	s := c.service
	s.lock.Lock()
	defer s.lock.Unlock()
	key := toKey(keys)
	myExisting := c.leases[key]
	realExisting := s.leases[key]
	if myExisting != nil {
		if myExisting != realExisting {
			// The lease expired, we just fail and don't try to re-acquire it
			// Otherwise it is possible another client acquired and released the lease in the meantime
			// so we should make sure this client knows that the lease was not valid for the whole time
			return false
		}
		if myExisting.After(time.Now()) {
			// We already hold the lease
			exp := time.Now().Add(ttl)
			c.leases[key] = &exp
			s.leases[key] = &exp
			return true
		}
		// The lease expired, we just fail and don't try to re-acquire it
		// Otherwise it is possible another client acquired and released the lease in the meantime
		// Unlikely as the ttl is the same, but still possible
		return false
	}
	if realExisting != nil && realExisting.After(time.Now()) {
		// Someone else holds the lease
		return false
	}
	// grab the lease
	exp := time.Now().Add(ttl)
	c.leases[key] = &exp
	s.leases[key] = &exp
	return true
}

func toKey(key []string) string {
	return strings.Join(key, "/")
}

type leaseClient struct {
	// A local record of all leases held by this client, my diverge if they are not renewed in time
	leases  map[string]*time.Time
	service *Service
}
