package state

import (
	"github.com/block/ftl/common/sha256"
)

type DeploymentArtefact struct {
	Digest     sha256.SHA256
	Path       string
	Executable bool
}
