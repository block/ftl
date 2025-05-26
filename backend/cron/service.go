package cron

import (
	"context"
	"fmt"
	"net/url"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/cron/observability"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/cron"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/statemachine"
	"github.com/block/ftl/internal/timelineclient"
)

type cronJob struct {
	module     string
	deployment key.Deployment
	verb       *schema.Verb
	cronmd     *schema.MetadataCronJob
	pattern    cron.Pattern
	next       time.Time
}

func (c cronJob) Key() string {
	return c.module + "." + c.verb.Name
}

func (c cronJob) String() string {
	desc := fmt.Sprintf("%s.%s (%s)", c.module, c.verb.Name, c.pattern)
	var next string
	if time.Until(c.next) > 0 {
		next = fmt.Sprintf(" (next run in %s)", time.Until(c.next))
	}
	return desc + next
}

type Config struct {
	Bind                  *url.URL        `help:"Address to bind to." env:"FTL_BIND" default:"http://127.0.0.1:8990"`
	SchemaServiceEndpoint *url.URL        `name:"ftl-endpoint" help:"Schema Service endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineEndpoint      *url.URL        `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
	Raft                  raft.RaftConfig `embed:"" prefix:"raft-"`
}

// Start the cron service. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config, eventSource *schemaeventsource.EventSource, client routing.CallClient, timelineClient *timelineclient.Client) error {
	logger := log.FromContext(ctx).Scope("cron")
	ctx = log.ContextWithLogger(ctx, logger)
	// Map of cron jobs for each module.
	cronJobs := map[string][]*cronJob{}
	// Cron jobs ordered by next execution.
	cronQueue := []*cronJob{}

	logger.Debugf("Starting FTL cron service")
	events := eventSource.Subscribe(ctx)

	var shard statemachine.Handle[struct{}, CronState, CronEvent]

	g, ctx := errgroup.WithContext(ctx)

	if config.Raft.DataDir == "" {
		shard = statemachine.NewLocalHandle(newStateMachine(ctx))
	} else {
		gctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM)
		defer cancel()

		clusterBuilder := raft.NewBuilder(&config.Raft)
		shard = raft.AddShard(gctx, clusterBuilder, 1, newStateMachine(ctx))
		cluster := clusterBuilder.Build(gctx)

		var rpcOpts []rpc.Option
		rpcOpts = append(rpcOpts, raft.RPCOption(cluster))
		g.Go(func() error {
			return errors.WithStack(rpc.Serve(ctx, config.Bind, rpcOpts...))
		})
	}

	logger.Debugf("Starting cron service")
	state := statemachine.NewSingleQueryHandle(shard, struct{}{})

	g.Go(func() error {
		for {
			next, ok := scheduleNext(cronQueue)
			var nextCh <-chan time.Time
			if ok {
				if next == 0 {
					// Execute immediately
					select {
					case <-ctx.Done():
						return errors.Wrap(ctx.Err(), "cron service stopped")
					default:
						if err := executeJob(ctx, state, client, cronQueue[0], timelineClient); err != nil {
							logger.Errorf(err, "Failed to execute job")
						}
						orderQueue(cronQueue)
						continue
					}
				}
				logger.Debugf("Next cron job scheduled in %s", next)
				nextCh = time.After(next)
			} else {
				logger.Debugf("No cron jobs scheduled")
			}
			select {
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "cron service stopped")

			case change := <-events:
				if err := updateCronJobs(ctx, cronJobs, change, timelineClient); err != nil {
					logger.Errorf(err, "Failed to update cron jobs")
					continue
				}
				cronQueue = rebuildQueue(cronJobs)

			// Execute scheduled cron job
			case <-nextCh:
				if err := executeJob(ctx, state, client, cronQueue[0], timelineClient); err != nil {
					logger.Errorf(err, "Failed to execute job")
					continue
				}
				orderQueue(cronQueue)
			}
		}
	})

	err := g.Wait()
	if err != nil {
		if ctx.Err() == nil {
			// startup failure if the context was not cancelled
			return errors.Wrap(err, "failed to start cron service")
		}
	}
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "cron service stopped")
	}
	return nil
}

func executeJob(ctx context.Context, state *statemachine.SingleQueryHandle[struct{}, CronState, CronEvent], client routing.CallClient, job *cronJob, timelineClient *timelineclient.Client) error {
	logger := log.FromContext(ctx).Scope("cron").Module(job.module)
	logger.Debugf("Executing cron job %s", job)

	view, err := state.View(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get job state")
	}

	lastExec, hasLast := view.LastExecutions[job.Key()]
	if hasLast && !job.next.After(lastExec) {
		logger.Debugf("Skipping already executed job %s", job.Key())
		return nil
	}

	nextRun, err := cron.Next(job.pattern, false)
	if err != nil {
		return errors.Wrap(err, "failed to calculate next run time")
	}

	event := CronEvent{
		JobKey:        job.Key(),
		ExecutedAt:    time.Now(),
		NextExecution: nextRun,
	}
	if err := state.Publish(ctx, event); err != nil {
		return errors.Wrap(err, "failed to claim job execution")
	}

	job.next = nextRun

	cronModel := model.CronJob{
		Key:           key.NewCronJobKey(job.module, job.verb.Name),
		Verb:          schema.Ref{Module: job.module, Name: job.verb.Name},
		Schedule:      job.pattern.String(),
		StartTime:     event.ExecutedAt,
		NextExecution: job.next,
	}

	timelineClient.Publish(ctx, timelineclient.CronScheduled{
		DeploymentKey: job.deployment,
		Verb:          schema.Ref{Module: job.module, Name: job.verb.Name},
		ScheduledAt:   job.next,
		Schedule:      job.pattern.String(),
	})

	observability.Cron.JobStarted(ctx, cronModel)
	if err := callCronVerb(ctx, client, job); err != nil {
		observability.Cron.JobFailed(ctx, cronModel)
		return errors.Wrap(err, "failed to execute cron job")
	}
	observability.Cron.JobSuccess(ctx, cronModel)
	return nil
}

func callCronVerb(ctx context.Context, verbClient routing.CallClient, cronJob *cronJob) error {
	logger := log.FromContext(ctx).Scope("cron").Module(cronJob.module)
	ref := schema.Ref{Module: cronJob.module, Name: cronJob.verb.Name}
	logger.Debugf("Calling cron job %s", cronJob)

	req := connect.NewRequest(&ftlv1.CallRequest{
		Verb:     ref.ToProto(),
		Body:     []byte(`{}`),
		Metadata: &ftlv1.Metadata{},
	})

	requestKey := key.NewRequestKey(key.OriginCron, schema.RefKey{Module: ref.Module, Name: ref.Name}.String())
	headers.SetRequestKey(req.Header(), requestKey)

	resp, err := verbClient.Call(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "%s: call to cron job failed", ref)
	}
	switch resp := resp.Msg.Response.(type) {
	default:
		return nil

	case *ftlv1.CallResponse_Error_:
		return errors.Errorf("%s: cron job failed: %s", ref, resp.Error.Message)
	}
}

func scheduleNext(cronQueue []*cronJob) (time.Duration, bool) {
	if len(cronQueue) == 0 {
		return 0, false
	}

	// If next execution is in the past, schedule immediately
	next := time.Until(cronQueue[0].next)
	if next < 0 {
		next = 0
	}

	return next, true
}

func updateCronJobs(ctx context.Context, cronJobs map[string][]*cronJob, change schema.Notification, timelineClient *timelineclient.Client) error {
	logger := log.FromContext(ctx).Scope("cron")

	// Track jobs before the update to detect changes
	oldJobs := make(map[string]*cronJob)
	for _, jobs := range cronJobs {
		for _, job := range jobs {
			oldJobs[job.Key()] = job
		}
	}

	switch change := change.(type) {
	case *schema.FullSchemaNotification:
		for _, module := range change.Schema.InternalModules() {
			logger.Debugf("Updated cron jobs for module %s", module.Name)
			moduleJobs, err := extractCronJobs(module)
			if err != nil {
				return errors.Wrap(err, "failed to extract cron jobs")
			}
			logger.Debugf("Adding %d cron jobs for module %s", len(moduleJobs), module.Name)
			cronJobs[module.Name] = moduleJobs

			// Publish timeline events for new or changed jobs
			publishNewOrChangedJobs(ctx, moduleJobs, oldJobs, timelineClient)
		}
	case *schema.ChangesetCommittedNotification:
		for _, removed := range change.Changeset.InternalRealm().RemovingModules {
			// These are modules that are properly deleted
			logger.Debugf("Removing cron jobs for module %s", removed.Name)
			delete(cronJobs, removed.Name)
		}
		for _, module := range change.Changeset.InternalRealm().Modules {
			logger.Debugf("Updated cron jobs for module %s", module.Name)
			moduleJobs, err := extractCronJobs(module)
			if err != nil {
				return errors.Wrap(err, "failed to extract cron jobs")
			}
			logger.Tracef("Adding %d cron jobs for module %s", len(moduleJobs), module)
			cronJobs[module.Name] = moduleJobs

			// Publish timeline events for new or changed jobs
			publishNewOrChangedJobs(ctx, moduleJobs, oldJobs, timelineClient)
		}
	default:
	}
	return nil
}

func publishNewOrChangedJobs(ctx context.Context, newJobs []*cronJob, oldJobs map[string]*cronJob, timelineClient *timelineclient.Client) {
	for _, job := range newJobs {
		oldJob, exists := oldJobs[job.Key()]
		// Publish event if job is new or if the schedule has changed
		if !exists || oldJob.pattern.String() != job.pattern.String() {
			timelineClient.Publish(ctx, timelineclient.CronScheduled{
				DeploymentKey: job.deployment,
				Verb:          schema.Ref{Module: job.module, Name: job.verb.Name},
				ScheduledAt:   job.next,
				Schedule:      job.pattern.String(),
			})
		}
	}
}

func orderQueue(queue []*cronJob) {
	sort.SliceStable(queue, func(i, j int) bool {
		return queue[i].next.Before(queue[j].next)
	})
}

func rebuildQueue(cronJobs map[string][]*cronJob) []*cronJob {
	queue := make([]*cronJob, 0, len(cronJobs)*2) // Assume 2 cron jobs per module.
	for _, jobs := range cronJobs {
		queue = append(queue, jobs...)
	}
	orderQueue(queue)
	return queue
}

func extractCronJobs(module *schema.Module) ([]*cronJob, error) {
	if module.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
		return nil, nil
	}
	cronJobs := []*cronJob{}
	for verb := range slices.FilterVariants[*schema.Verb](module.Decls) {
		cronmd, ok := slices.FindVariant[*schema.MetadataCronJob](verb.Metadata)
		if !ok {
			continue
		}
		pattern, err := cron.Parse(cronmd.Cron)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", cronmd.Pos)
		}
		next, err := cron.Next(pattern, false)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", cronmd.Pos)
		}
		cronJobs = append(cronJobs, &cronJob{
			module:     module.Name,
			deployment: module.Runtime.Deployment.DeploymentKey,
			verb:       verb,
			cronmd:     cronmd,
			pattern:    pattern,
			next:       next,
		})
	}
	return cronJobs, nil
}
