package adminpb

import (
	"github.com/block/ftl/backend/controller/state"
)

func ArtefactToProto(artefact *state.DeploymentArtefact) *DeploymentArtefact {
	return &DeploymentArtefact{
		Path:       artefact.Path,
		Executable: artefact.Executable,
		Digest:     artefact.Digest[:],
	}
}
