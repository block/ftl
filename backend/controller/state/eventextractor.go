package state

import (
	"iter"

	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/internal/iterops"
)

// EventExtractor calculates controller events from changes to the state.
func EventExtractor(diff tuple.Pair[SchemaState, SchemaState]) iter.Seq[SchemaEvent] {
	var events []SchemaEvent

	previous := diff.A
	current := diff.B

	previousAll := previous.GetDeployments()
	for _, deployment := range current.GetDeployments() {
		pd, ok := previousAll[deployment.Key.String()]
		if !ok {
			events = append(events, &DeploymentCreatedEvent{
				Module:    deployment.Module,
				Key:       deployment.Key,
				CreatedAt: deployment.CreatedAt,
				Schema:    deployment.Schema,
				Language:  deployment.Language,
			})
		} else if !pd.Schema.Equals(deployment.Schema) {
			events = append(events, &DeploymentSchemaUpdatedEvent{
				Key:    deployment.Key,
				Schema: deployment.Schema,
			})
		}
	}

	currentActive := current.GetActiveDeployments()
	currentAll := current.GetDeployments()
	currentModules := map[string]bool{}
	for _, deployment := range currentAll {
		currentModules[deployment.Module] = true
	}
	for _, deployment := range previous.GetActiveDeployments() {
		if _, ok := currentActive[deployment.Key.String()]; !ok {
			_, ok2 := currentModules[deployment.Module]
			events = append(events, &DeploymentDeactivatedEvent{
				Key:           deployment.Key,
				ModuleRemoved: !ok2,
			})
		}
	}
	return iterops.Const(events...)
}
