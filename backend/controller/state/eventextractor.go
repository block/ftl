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
	for key, deployment := range current.GetDeployments() {
		pd, ok := previousAll[key]
		if !ok {
			events = append(events, &DeploymentCreatedEvent{
				Key:       key,
				CreatedAt: deployment.GetRuntime().GetDeployment().GetCreatedAt(),
				Schema:    deployment,
			})
		} else if !pd.Equals(deployment) {
			events = append(events, &DeploymentSchemaUpdatedEvent{
				Key:    key,
				Schema: deployment,
			})
		}
	}

	currentActive := current.GetActiveDeployments()
	currentAll := current.GetDeployments()
	currentModules := map[string]bool{}
	for _, deployment := range currentAll {
		currentModules[deployment.Name] = true
	}
	for key, deployment := range previous.GetActiveDeployments() {
		if _, ok := currentActive[key]; !ok {
			_, ok2 := currentModules[deployment.Name]
			events = append(events, &DeploymentDeactivatedEvent{
				Key:           key,
				ModuleRemoved: !ok2,
			})
		}
	}
	return iterops.Const(events...)
}
