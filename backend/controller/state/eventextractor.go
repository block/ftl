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
		dk := deployment.Schema.GetRuntime().GetDeployment().GetDeploymentKey()
		pd, ok := previousAll[dk]
		if !ok {
			events = append(events, &DeploymentCreatedEvent{
				Key:       dk,
				CreatedAt: deployment.Schema.GetRuntime().GetDeployment().GetCreatedAt(),
				Schema:    deployment.Schema,
			})
		} else if !pd.Schema.Equals(deployment.Schema) {
			events = append(events, &DeploymentSchemaUpdatedEvent{
				Key:    dk,
				Schema: deployment.Schema,
			})
		}
	}

	currentActive := current.GetActiveDeployments()
	currentAll := current.GetDeployments()
	currentModules := map[string]bool{}
	for _, deployment := range currentAll {
		currentModules[deployment.Schema.Name] = true
	}
	for _, deployment := range previous.GetActiveDeployments() {
		dk := deployment.Schema.GetRuntime().GetDeployment().GetDeploymentKey()
		if _, ok := currentActive[dk]; !ok {
			_, ok2 := currentModules[deployment.Schema.Name]
			events = append(events, &DeploymentDeactivatedEvent{
				Key:           dk,
				ModuleRemoved: !ok2,
			})
		}
	}
	return iterops.Const(events...)
}
