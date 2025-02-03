package schemaservice

import (
	"iter"
	"slices"
	"time"

	"github.com/alecthomas/types/tuple"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/key"
)

// EventExtractor calculates controller events from changes to the state.
func EventExtractor(diff tuple.Pair[SchemaState, SchemaState]) iter.Seq[schema.Event] {
	var events []schema.Event

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
				for i := range changeset.Modules {
					pd := pc.Modules[i]
					deployment := changeset.Modules[i]
					if !pd.Equals(deployment) {
						handledDeployments[deployment.Runtime.Deployment.DeploymentKey] = true
						// TODO: this seems super inefficient, we should not need to do equality checks on every deployment
						events = append(events, &schema.DeploymentSchemaUpdatedEvent{
							Key:       deployment.Runtime.Deployment.DeploymentKey,
							Schema:    deployment,
							Changeset: changeset.Key,
						})
					} else {
						handledDeployments[deployment.Runtime.Deployment.DeploymentKey] = true
					}
				}
			}
			// Commit final state of changeset
			if changeset.State == schema.ChangesetStateCommitted && pc.State != schema.ChangesetStateCommitted {
				events = append(events, &schema.ChangesetCommittedEvent{
					Key: changeset.Key,
				})
				for _, deployment := range changeset.Modules {
					events = append(events, &schema.DeploymentActivatedEvent{
						Key:         deployment.Runtime.Deployment.DeploymentKey,
						MinReplicas: 1,
						ActivatedAt: time.Now(),
						Changeset:   &changeset.Key,
					})
				}
			} else if changeset.State == schema.ChangesetStateFailed && pc.State != schema.ChangesetStateFailed {
				events = append(events, &schema.ChangesetFailedEvent{
					Key:   changeset.Key,
					Error: changeset.Error,
				})
			}
		} else {
			// New changeset and associated modules
			events = append(events, &schema.ChangesetCreatedEvent{
				Changeset: changeset,
			})
			// Find new deployments from the changeset
			for _, deployment := range changeset.Modules {
				// changeset is always a new deployment
				events = append(events, &schema.DeploymentCreatedEvent{
					Key:       deployment.Runtime.Deployment.DeploymentKey,
					Schema:    deployment,
					Changeset: &changeset.Key,
				})
				handledDeployments[deployment.Runtime.Deployment.DeploymentKey] = true
			}
		}
	}

	for key, deployment := range current.GetAllActiveDeployments() {
		if handledDeployments[key] {
			// Already handled in the changeset
			continue
		}
		pd, ok := previousAllDeployments[key]
		if !ok {
			// We have lost the changeset that created this, this should only happen
			// if the changeset was deleted as part of raft cleanup.
			events = append(events, &schema.DeploymentCreatedEvent{
				Key:    key,
				Schema: deployment,
			})
		} else if !pd.Equals(deployment) {
			// TODO: this seems super inefficient, we should not need to do equality checks on every deployment
			events = append(events, &schema.DeploymentSchemaUpdatedEvent{
				Key:    key,
				Schema: deployment,
			})
		}
	}

	currentActive := current.GetCanonicalDeployments()
	currentAll := current.GetDeployments()
	currentModules := map[string]bool{}
	for _, deployment := range currentAll {
		currentModules[deployment.Name] = true
	}
	for key, deployment := range previous.GetCanonicalDeployments() {
		if _, ok := currentActive[key]; !ok {
			_, ok2 := currentModules[deployment.Name]
			events = append(events, &schema.DeploymentDeactivatedEvent{
				Key:           key,
				ModuleRemoved: !ok2,
			})
		}
	}
	return iterops.Const(events...)
}
