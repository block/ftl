package raft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/lni/dragonboat/v4"

	raftpb "github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/log"
)

// AddMember to the cluster. This needs to be called on an existing running cluster member,
// before the new member is started.
func (c *Cluster) AddMember(ctx context.Context, req *connect.Request[raftpb.AddMemberRequest]) (*connect.Response[raftpb.AddMemberResponse], error) {
	logger := log.FromContext(ctx).Scope("raft")

	shards := req.Msg.ShardIds
	replicaID := req.Msg.ReplicaId
	address := req.Msg.Address

	logger.Infof("Adding member %s to shard %d on replica %d", address, shards, replicaID)

	for _, shardID := range shards {
		if err := c.withRetry(ctx, shardID, replicaID, func(ctx context.Context) error {
			logger.Debugf("Requesting add replica to shard %d on replica %d", shardID, replicaID)
			return c.nh.SyncRequestAddReplica(ctx, shardID, replicaID, address, 0)
		}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
			if errors.Is(err, dragonboat.ErrRejected) {
				// The request can be rejected if the member is already in the cluster.
				// Check if the member in the cluster matches the requested address.
				ok, err := c.isMember(ctx, shardID, replicaID, address)
				if err != nil {
					return nil, fmt.Errorf("failed to check if member is already in cluster: %w", err)
				}
				if !ok {
					return nil, fmt.Errorf("failed to add member not in cluster: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to add member: %w", err)
			}
		}
	}
	return connect.NewResponse(&raftpb.AddMemberResponse{}), nil
}

func (c *Cluster) isMember(ctx context.Context, shardID uint64, replicaID uint64, address string) (bool, error) {
	timeout := time.Duration(c.config.ElectionRTT) * c.config.RTT
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	membership, err := c.nh.SyncGetShardMembership(ctx, shardID)
	if err != nil {
		return false, fmt.Errorf("failed to get cluster info: %w", err)
	}

	for id, addr := range membership.Nodes {
		if id == replicaID && addr == address {
			return true, nil
		}
	}
	return false, nil
}

func (c *Cluster) RemoveMember(ctx context.Context, req *connect.Request[raftpb.RemoveMemberRequest]) (*connect.Response[raftpb.RemoveMemberResponse], error) {
	logger := log.FromContext(ctx).Scope("raft")
	logger.Infof("Request to remove member %d from shards %v on replica %d", req.Msg.ReplicaId, req.Msg.ShardIds, req.Msg.ReplicaId)

	for _, shardID := range req.Msg.ShardIds {
		if err := c.removeShardMember(ctx, shardID, req.Msg.ReplicaId); err != nil {
			return nil, fmt.Errorf("failed to remove member from shard %d: %w", shardID, err)
		}
	}

	return connect.NewResponse(&raftpb.RemoveMemberResponse{}), nil
}

// Ping the cluster.
func (c *Cluster) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// removeShardMember from the given shard. This removes the given member from the membership group
// and blocks until the change has been committed
func (c *Cluster) removeShardMember(ctx context.Context, shardID uint64, replicaID uint64) error {
	logger := log.FromContext(ctx).Scope("raft")
	logger.Infof("Removing replica %d from shard %d", shardID, replicaID)

	if err := c.withRetry(ctx, shardID, replicaID, func(ctx context.Context) error {
		logger.Debugf("Requesting delete replica from shard %d on replica %d", shardID, replicaID)
		return c.nh.SyncRequestDeleteReplica(ctx, shardID, replicaID, 0)
	}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
		// This can happen if the cluster is shutting down and no longer has quorum.
		logger.Warnf("Removing replica %d from shard %d failed: %s", replicaID, shardID, err)
		return err
	}
	return nil
}
