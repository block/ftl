package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/internal/download"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type downloadCmd struct {
	Dest       string    `short:"d" help:"Destination directory." default:"."`
	Deployment ulid.ULID `help:"Deployment to download." arg:""`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.ControlPlaneServiceClient) error {
	return download.Artefacts(ctx, client, d.Deployment, d.Dest)
}

func (d *downloadCmd) getLocalArtefacts() ([]*ftlv1.DeploymentArtefact, error) {
	haveArtefacts := []*ftlv1.DeploymentArtefact{}
	dest, err := filepath.Abs(d.Dest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		sum, err := sha256.SumFile(path)
		if err != nil {
			return errors.WithStack(err)
		}

		relPath, err := filepath.Rel(dest, path)
		if err != nil {
			return errors.WithStack(err)
		}
		haveArtefacts = append(haveArtefacts, &ftlv1.DeploymentArtefact{
			Path:       relPath,
			Digest:     sum.String(),
			Executable: info.Mode()&0111 != 0,
		})
		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return haveArtefacts, nil
}
