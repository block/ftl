package schemaservice

import (
	"fmt"
	"iter"
	"slices"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/tuple"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
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
	slices.SortFunc(allChangesets, func(a, b *schema.Changeset) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	for _, changeset := range allChangesets {
		pc, ok := previousAllChangesets[changeset.Key]
		if !ok {
			events = append(events, &ChangesetCreatedEvent{
				Changeset: changeset,
			})
			continue
		}

		// Find changes in modules
		prevModules := islices.Reduce(pc.Modules, map[key.Deployment]*schema.Module{}, func(acc map[key.Deployment]*schema.Module, m *schema.Module) map[key.Deployment]*schema.Module {
			acc[m.Runtime.Deployment.DeploymentKey] = m
			return acc
		})
		for _, module := range changeset.Modules {
			prevModule, ok := prevModules[module.Runtime.Deployment.DeploymentKey]
			if !ok {
				panic(fmt.Sprintf("can not create deployment %s in %s", module.Runtime.Deployment.DeploymentKey, changeset.Key))
			}
			if !prevModule.Equals(module) {
				events = append(events, &DeploymentSchemaUpdatedEvent{
					Key:       module.Runtime.Deployment.DeploymentKey,
					Schema:    module,
					Changeset: optional.Some(changeset.Key),
				})
			}
		}

		// Commit final state of changeset
		if changeset.State == schema.ChangesetStateCommitted && pc.State != schema.ChangesetStateCommitted {
			events = append(events, &ChangesetCommittedEvent{
				Key: changeset.Key,
			})
			// Maintain previous state to avoid unnecessary events
			for _, module := range changeset.Modules {
				previousAllDeployments[module.Runtime.Deployment.DeploymentKey] = module
			}
		} else if changeset.State == schema.ChangesetStateFailed && pc.State != schema.ChangesetStateFailed {
			events = append(events, &ChangesetFailedEvent{
				Key:   changeset.Key,
				Error: changeset.Error,
			})
		}
	}

	for key, deployment := range current.GetDeployments() {
		pd, ok := previousAllDeployments[key]
		if !ok {
			events = append(events, &schema.DeploymentCreatedEvent{
				Key:    key,
				Schema: deployment,
			})
		} else if !pd.Equals(deployment) {
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
