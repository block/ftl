package schemaservice

import (
	"iter"
	"slices"

	"github.com/alecthomas/types/tuple"
	"golang.org/x/exp/maps"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
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
	previousAllDeployments := previous.GetDeployments()
	previousAllChangesets := previous.GetChangesets()
	allChangesets := maps.Values(current.GetChangesets())
	handledDeployments := map[key.Deployment]bool{}
	slices.SortFunc(allChangesets, func(a, b *schema.Changeset) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	for _, changeset := range allChangesets {
		pc, ok := previousAllChangesets[changeset.Key]
		if ok {

			if changeset.ModulesAreCanonical() {
				csEvents := current.changesetEvents[changeset.Key]
				prEvents := previous.changesetEvents[changeset.Key]
				for _, dep := range changeset.Modules {
					handledDeployments[dep.Runtime.Deployment.DeploymentKey] = true
				}
				// Use the event list length to check for changes, and send updated events
				if len(csEvents) > len(prEvents) {
					csName := changeset.Key.String()
					changedDeployments := map[key.Deployment]bool{}
					for _, event := range csEvents[len(prEvents):] {
						changedDeployments[event.DeploymentKey()] = true
					}
					for _, e := range changeset.Modules {
						if changedDeployments[e.Runtime.Deployment.DeploymentKey] {
							events = append(events, &ftlv1.PullSchemaResponse{
								Event: &ftlv1.PullSchemaResponse_DeploymentUpdated_{
									DeploymentUpdated: &ftlv1.PullSchemaResponse_DeploymentUpdated{
										Changeset: &csName,
										Schema:    e.ToProto(),
									},
								},
							})
						}
					}
				}
			} else if changeset.State == schema.ChangesetStateCommitted && pc.State != schema.ChangesetStateCommitted {
				// New changeset and associated modules
				events = append(events, &ftlv1.PullSchemaResponse{
					Event: &ftlv1.PullSchemaResponse_ChangesetCommitted_{
						ChangesetCommitted: &ftlv1.PullSchemaResponse_ChangesetCommitted{
							Changeset: changeset.ToProto(),
						},
					},
				})
				for _, removing := range pc.RemovingModules {
					handledDeployments[removing.Runtime.Deployment.DeploymentKey] = true
				}
				// These modules were added as part of the committed changeset, we don't want to treat them as new ones
				for _, dep := range changeset.Modules {
					handledDeployments[dep.Runtime.Deployment.DeploymentKey] = true
				}
			} else if changeset.State == schema.ChangesetStateFailed && pc.State != schema.ChangesetStateFailed {
				events = append(events, &ftlv1.PullSchemaResponse{
					Event: &ftlv1.PullSchemaResponse_ChangesetFailed_{
						ChangesetFailed: &ftlv1.PullSchemaResponse_ChangesetFailed{
							Key: changeset.Key.String(),
						},
					},
				})
			}
		} else {
			// New changeset and associated modules
			events = append(events, &ftlv1.PullSchemaResponse{
				Event: &ftlv1.PullSchemaResponse_ChangesetCreated_{
					ChangesetCreated: &ftlv1.PullSchemaResponse_ChangesetCreated{
						Changeset: changeset.ToProto(),
					},
				},
			})
		}
	}
	disappearedDeployments := map[key.Deployment]bool{}
	for _, v := range previousAllDeployments {
		disappearedDeployments[v.Runtime.Deployment.DeploymentKey] = true
	}

	for dplKey, deployment := range current.GetAllActiveDeployments() {
		delete(disappearedDeployments, dplKey)
		if handledDeployments[dplKey] {
			// Already handled in the changeset
			continue
		}
		_, ok := previousAllDeployments[dplKey]
		if !ok {
			// We have lost the changeset that created this, this should only happen
			// if the changeset was deleted as part of raft cleanup.
			events = append(events, &ftlv1.PullSchemaResponse{
				Event: &ftlv1.PullSchemaResponse_DeploymentCreated_{
					DeploymentCreated: &ftlv1.PullSchemaResponse_DeploymentCreated{
						Schema: deployment.ToProto(),
					},
				},
			})
		} else {
			// Check for runtime events
			csEvents := current.deploymentEvents[deployment.Name] // These are keyed by module name, but we know this is the same deployment
			prEvents := previous.deploymentEvents[deployment.Name]
			// Use the event list length to check for changes, and send updated events
			if len(csEvents) > len(prEvents) {
				events = append(events, &ftlv1.PullSchemaResponse{
					Event: &ftlv1.PullSchemaResponse_DeploymentUpdated_{
						DeploymentUpdated: &ftlv1.PullSchemaResponse_DeploymentUpdated{
							Schema: deployment.ToProto(),
						},
					},
				})
			}
		}
	}
	for k := range disappearedDeployments {
		deplKey := k.String()
		events = append(events, &ftlv1.PullSchemaResponse{
			Event: &ftlv1.PullSchemaResponse_DeploymentRemoved_{
				DeploymentRemoved: &ftlv1.PullSchemaResponse_DeploymentRemoved{
					Key:        &deplKey,
					ModuleName: k.Payload.Module,
				},
			},
		})
	}

	return iterops.Const(events...)
}
