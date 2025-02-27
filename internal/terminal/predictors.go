package terminal

import (
	"github.com/posener/complete"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func Predictors(view *schemaeventsource.View) map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"verbs":         &declPredictor[*schema.Verb]{view: *view},
		"configs":       &declPredictor[*schema.Config]{view: *view},
		"secrets":       &declPredictor[*schema.Secret]{view: *view},
		"databases":     &declPredictor[*schema.Database]{view: *view},
		"subscriptions": &subscriptionsPredictor{view: *view},
		"modules":       &modulePredictor{view: *view},
		"deployments":   &deploymentsPredictor{view: *view},
		"log-level":     &logLevelPredictor{},
	}
}

type declPredictor[d schema.Decl] struct {
	view schemaeventsource.View
}

func (p *declPredictor[d]) Predict(args complete.Args) []string {
	sch := p.view.GetCanonical()
	ret := []string{}
	for _, module := range sch.Modules {
		for _, dec := range module.Decls {
			if _, ok := dec.(d); ok {
				ref := schema.Ref{Module: module.Name, Name: dec.GetName()}
				ret = append(ret, ref.String())
			}
		}
	}
	return ret
}

type subscriptionsPredictor struct {
	view schemaeventsource.View
}

func (p *subscriptionsPredictor) Predict(args complete.Args) []string {
	sch := p.view.GetCanonical()
	ret := []string{}
	for _, module := range sch.Modules {
		for _, dec := range module.Decls {
			verb, ok := dec.(*schema.Verb)
			if !ok {
				continue
			}
			_, ok = slices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
			if !ok {
				continue
			}
			ref := schema.Ref{Module: module.Name, Name: dec.GetName()}
			ret = append(ret, ref.String())
		}
	}
	return ret
}

type modulePredictor struct {
	view schemaeventsource.View
}

func (v *modulePredictor) Predict(args complete.Args) []string {
	sch := v.view.GetCanonical()
	ret := []string{}
	for _, module := range sch.Modules {
		ret = append(ret, module.Name)
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
