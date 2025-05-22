package artefacts

import (
	"os"
	"path/filepath"
	"sort"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
)

var _ DeploymentArtefactProvider = (*localDeploymentProvider)(nil)

type localDeploymentProvider struct {
	path string
}

// Provide implements DeploymentArtefactProvider.
func (l *localDeploymentProvider) Provide() (string, error) {
	return l.path, nil
}

// manageDeploymentDirectory ensures the deployment directory exists and removes old deployments.
func manageDeploymentDirectory(logger *log.Logger, deploymentDir string) error {
	logger.Debugf("Deployment directory: %s", deploymentDir)
	err := os.MkdirAll(deploymentDir, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment directory")
	}

	// Clean up old deployments.
	modules, err := os.ReadDir(deploymentDir)
	if err != nil {
		return errors.Wrap(err, "failed to read deployment directory")
	}

	for _, module := range modules {
		if !module.IsDir() {
			continue
		}

		moduleDir := filepath.Join(deploymentDir, module.Name())
		deployments, err := os.ReadDir(moduleDir)
		if err != nil {
			return errors.Wrap(err, "failed to read module directory")
		}

		stats, err := slices.MapErr(deployments, func(d os.DirEntry) (os.FileInfo, error) {
			return errors.WithStack2(d.Info())
		})
		if err != nil {
			return errors.Wrap(err, "failed to stat deployments")
		}

		// Sort deployments by modified time, remove anything past the history limit.
		sort.Slice(deployments, func(i, j int) bool {
			return stats[i].ModTime().After(stats[j].ModTime())
		})

		for _, deployment := range deployments {
			old := filepath.Join(moduleDir, deployment.Name())
			logger.Debugf("Removing old deployment: %s", old)

			err := os.RemoveAll(old)
			if err != nil {
				// This is not a fatal error, just log it.
				logger.Errorf(err, "Failed to remove old deployment: %s", deployment.Name())
			}
		}
	}

	return nil
}
