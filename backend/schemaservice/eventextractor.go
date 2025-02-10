package schemaservice

import (
	"iter"
	"slices"

	"github.com/alecthomas/types/tuple"
	"golang.org/x/exp/maps"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/key"
)

// EventExtractor calculates controller events from changes to the state.
func EventExtractor(diff tuple.Pair[SchemaState, SchemaState]) iter.Seq[*ftlv1.PullSchemaResponse] {
	var events []*ftlv1.PullSchemaResponse

	previous := diff.A
	current := diff.B

	// previousAllDeployments will be updated with committed changesets so that we only send relevant
	// DeploymentCreatedEvent and DeploymentSchemaUpdatedEvent events after changeset events have been processed by the receiver.
	previousAllDeployments := previous.GetCanonicalDeployments()
	previousAllChangesets := previous.GetChangesets()
	allChangesets := maps.Values(current.GetChangesets())
	handledDeployments := map[key.Deployment]bool{}

	for _, cs := range current.archivedChangesets {
		// Check if there are newly archived changesets
		if len(previous.archivedChangesets) > 0 {
			if previous.archivedChangesets[0].Key == cs.Key {
				break
			}
		}
		allChangesets = append(allChangesets, cs)
	}

	slices.SortFunc(allChangesets, func(a, b *schema.Changeset) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})

	disappearedDeployments := map[key.Deployment]bool{}
	for _, v := range previousAllDeployments {
		disappearedDeployments[v.Runtime.Deployment.DeploymentKey] = true
	}
	for _, changeset := range allChangesets {
		pc, ok := previousAllChangesets[changeset.Key]
		if ok {

			// Check for runtime changes in the changeset
			csEvents := current.changesetEvents[changeset.Key]
			prEvents := previous.changesetEvents[changeset.Key]
			for _, dep := range changeset.OwnedModules() {
				handledDeployments[dep.Runtime.Deployment.DeploymentKey] = true
				delete(disappearedDeployments, dep.Runtime.Deployment.DeploymentKey)
			}
			// Use the event list length to check for changes, and send updated events
			if len(csEvents) > len(prEvents) {
				for _, event := range csEvents[len(prEvents):] {
					e := &schema.DeploymentRuntimeNotification{
						Changeset: &changeset.Key,
						Payload:   event.Payload,
					}
					events = append(events, &ftlv1.PullSchemaResponse{
						Event: &schemapb.Notification{
							Value: &schemapb.Notification_DeploymentRuntimeNotification{
								DeploymentRuntimeNotification: e.ToProto()},
						},
					})
				}
			}
			if pc.State != changeset.State {
				// State changes
				// Note that we use larger than or equal to here, as we want to send the events for the state changes
				// even if we missed a transition state
				if changeset.State == schema.ChangesetStateRollingBack || changeset.State == schema.ChangesetStateFailed {
					// Fail conditions first, which means in the following else block we just have to deal with the happy path
					if changeset.State >= schema.ChangesetStateRollingBack && pc.State != schema.ChangesetStateRollingBack {
						notification := &schema.ChangesetRollingBackNotification{
							Changeset: changeset,
							Error:     changeset.Error,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetRollingBackNotification{ChangesetRollingBackNotification: notification.ToProto()}},
						})
					}
					if changeset.State == schema.ChangesetStateFailed {
						notification := &schema.ChangesetFailedNotification{
							Key:   changeset.Key,
							Error: changeset.Error,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetFailedNotification{ChangesetFailedNotification: notification.ToProto()}},
						})
					}
				} else {

					if changeset.State >= schema.ChangesetStatePrepared && pc.State < schema.ChangesetStatePrepared {
						// New changeset and associated modules
						notification := &schema.ChangesetPreparedNotification{
							Key: changeset.Key,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetPreparedNotification{ChangesetPreparedNotification: notification.ToProto()}},
						})
					}
					if changeset.State >= schema.ChangesetStateCommitted && pc.State < schema.ChangesetStateCommitted {
						// New changeset and associated modules
						notification := &schema.ChangesetCommittedNotification{
							Changeset: changeset,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetCommittedNotification{ChangesetCommittedNotification: notification.ToProto()}},
						})
						for _, removing := range pc.RemovingModules {
							handledDeployments[removing.Runtime.Deployment.DeploymentKey] = true
						}
						// These modules were added as part of the committed changeset, we don't want to treat them as new ones
						for _, dep := range changeset.Modules {
							handledDeployments[dep.Runtime.Deployment.DeploymentKey] = true
						}
					}
					if changeset.State >= schema.ChangesetStateDrained && pc.State < schema.ChangesetStateDrained {
						// New changeset and associated modules
						notification := &schema.ChangesetDrainedNotification{
							Key: changeset.Key,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetDrainedNotification{ChangesetDrainedNotification: notification.ToProto()}},
						})
					}
					if changeset.State >= schema.ChangesetStateFinalized && pc.State < schema.ChangesetStateFinalized {
						// New changeset and associated modules
						notification := &schema.ChangesetFinalizedNotification{
							Key: changeset.Key,
						}
						events = append(events, &ftlv1.PullSchemaResponse{
							Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetFinalizedNotification{ChangesetFinalizedNotification: notification.ToProto()}},
						})
					}
				}

			} else if changeset.State == schema.ChangesetStateFailed && pc.State != schema.ChangesetStateFailed {
				notification := &schema.ChangesetFailedNotification{
					Key:   changeset.Key,
					Error: changeset.Error,
				}
				events = append(events, &ftlv1.PullSchemaResponse{
					Event: &schemapb.Notification{Value: &schemapb.Notification_ChangesetFailedNotification{ChangesetFailedNotification: notification.ToProto()}},
				})
			}
		} else {
			e := &schema.ChangesetCreatedNotification{
				Changeset: changeset,
			}
			events = append(events, &ftlv1.PullSchemaResponse{
				Event: &schemapb.Notification{
					Value: &schemapb.Notification_ChangesetCreatedNotification{
						ChangesetCreatedNotification: e.ToProto(),
					},
				}})
		}
	}

	for dplKey, deployment := range current.GetCanonicalDeployments() {
		delete(disappearedDeployments, dplKey)
		if handledDeployments[dplKey] {
			// Already handled in the changeset
			continue
		}
		_, ok := previousAllDeployments[dplKey]
		if !ok {
			// We have lost the changeset that created this, this should only happen
			// if the changeset was deleted as part of raft cleanup.
			return sendFullSchema(&current)
		} else {
			// Check for runtime events
			csEvents := current.deploymentEvents[deployment.Name] // These are keyed by module name, but we know this is the same deployment
			prEvents := previous.deploymentEvents[deployment.Name]
			// Use the event list length to check for changes, and send updated events
			if len(csEvents) > len(prEvents) {
				for _, event := range csEvents[len(prEvents):] {
					e := &schema.DeploymentRuntimeNotification{
						Payload: event.Payload,
					}
					events = append(events, &ftlv1.PullSchemaResponse{
						Event: &schemapb.Notification{
							Value: &schemapb.Notification_DeploymentRuntimeNotification{
								DeploymentRuntimeNotification: e.ToProto()},
						},
					})
				}
			}
		}
	}
	if len(disappearedDeployments) > 0 {
		// We have lost the changeset that created this, this should only happen rarely, just sent the full schema
		return sendFullSchema(&current)
	}

	return iterops.Const(events...)
}

func sendFullSchema(current *SchemaState) iter.Seq[*ftlv1.PullSchemaResponse] {
	notification := &schema.FullSchemaNotification{
		Schema:     &schema.Schema{Modules: current.GetCanonicalDeploymentSchemas()},
		Changesets: maps.Values(current.GetChangesets()),
	}
	full := &ftlv1.PullSchemaResponse{Event: &schemapb.Notification{Value: &schemapb.Notification_FullSchemaNotification{
		FullSchemaNotification: notification.ToProto(),
	}}}
	// Just return the initial schema event, we don't need any more as this has all the info
	return iterops.Const(full)
}
