package observability

import (
	"github.com/alecthomas/errors"
)

var (
	Runner     *RunnerMetrics
	Deployment *DeploymentMetrics
)

func init() {
	var errs error
	var err error

	Runner, err = initRunnerMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)

	if errs != nil {
		panic(errors.Wrap(err, "could not initialize runner metrics"))
	}
}

func wrapErr(signalName string, err error) error {
	return errors.Wrapf(err, "failed to create %q signal", signalName)
}
