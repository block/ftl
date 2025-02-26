package terminal

import (
	"github.com/posener/complete"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func Predictors(view *schemaeventsource.View) map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"verbs":       &verbPredictor{view: *view},
		"deployments": &deploymentsPredictor{view: *view},
		"log-level":   &logLevelPredictor{},
	}
}

type verbPredictor struct {
	view schemaeventsource.View
}

func (v *verbPredictor) Predict(args complete.Args) []string {
	sch := v.view.GetCanonical()
	ret := []string{}
	for _, module := range sch.Modules {
		for _, dec := range module.Decls {
			if verb, ok := dec.(*schema.Verb); ok {
				ref := schema.Ref{Module: module.Name, Name: verb.Name}
				ret = append(ret, ref.String())
			}
		}
	}
	return ret
}

type deploymentsPredictor struct {
	view schemaeventsource.View
}

func (v *deploymentsPredictor) Predict(args complete.Args) []string {
	sch := v.view.GetCanonical()
	ret := []string{}
	for _, module := range sch.Modules {
		if module.Runtime == nil || module.Runtime.Deployment == nil {
			continue
		}
		ret = append(ret, module.Runtime.Deployment.DeploymentKey.String())
	}
	return ret
}

type logLevelPredictor struct {
}

func (v *logLevelPredictor) Predict(args complete.Args) []string {
	return log.LevelStrings()
}
