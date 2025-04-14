package download

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/controller/artefacts"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// Artefacts downloads artefacts for a deployment from the Controller.
func Artefacts(ctx context.Context, client adminpbconnect.AdminServiceClient, key key.Deployment, dest string) error {
	logger := log.FromContext(ctx)
	stream, err := client.GetDeploymentArtefacts(ctx, connect.NewRequest(&adminpb.GetDeploymentArtefactsRequest{
		DeploymentKey: key.String(),
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	start := time.Now()
	count := 0
	var digest []byte
	var w *os.File
	for stream.Receive() {
		msg := stream.Msg()
		artefact := msg.Artefact
		if !bytes.Equal(digest, artefact.Digest) {
			if w != nil {
				w.Close()
			}
			count++
			if !filepath.IsLocal(artefact.Path) {
				return errors.Errorf("path %q is not local", artefact.Path)
			}
			logger.Debugf("Downloading %s", filepath.Join(dest, artefact.Path))
			err = os.MkdirAll(filepath.Join(dest, filepath.Dir(artefact.Path)), 0700)
			if err != nil {
				return errors.WithStack(err)
			}
			var mode os.FileMode = 0600
			if artefact.Executable {
				mode = 0700
			}
			w, err = os.OpenFile(filepath.Join(dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return errors.WithStack(err)
			}
			digest = artefact.Digest
		}

		if _, err := w.Write(msg.Chunk); err != nil {
			_ = w.Close()
			return errors.WithStack(err)
		}
	}
	if w != nil {
		w.Close()
	}
	logger.Debugf("Downloaded %d artefacts in %s", count, time.Since(start))
	return errors.WithStack(stream.Err())
}

// ArtefactsFromOCI downloads artefacts for a deployment from an OCI registry.
func ArtefactsFromOCI(ctx context.Context, client ftlv1connect.SchemaServiceClient, key key.Deployment, dest string, service *artefacts.OCIArtefactService) error {
	logger := log.FromContext(ctx)
	response, err := client.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{
		DeploymentKey: key.String(),
	}))
	if err != nil {
		return errors.Wrapf(err, "failed to get deployment when downloading artifacts %q", key)
	}
	start := time.Now()
	count := 0
	for _, metadata := range response.Msg.Schema.Metadata {
		if artefact := metadata.GetArtefact(); artefact != nil {
			parseSHA256, err := sha256.ParseSHA256(artefact.Digest)
			if err != nil {
				return errors.Wrapf(err, "failed to parse SHA256 %q", artefact.Digest)
			}
			res, err := service.Download(ctx, parseSHA256)
			if err != nil {
				return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
			}
			count++
			if !filepath.IsLocal(artefact.Path) {
				return errors.Errorf("path %q is not local", artefact.Path)
			}
			logger.Debugf("Downloading %s", filepath.Join(dest, artefact.Path))
			err = os.MkdirAll(filepath.Join(dest, filepath.Dir(artefact.Path)), 0700)
			if err != nil {
				return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
			}
			var mode os.FileMode = 0600
			if artefact.Executable {
				mode = 0700
			}
			w, err := os.OpenFile(filepath.Join(dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
			}
			defer w.Close()
			buf := make([]byte, 1024)
			read := 0
			for {
				read, err = res.Read(buf)
				if read > 0 {
					_, e2 := w.Write(buf[:read])
					if e2 != nil {
						return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
					}
				}
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
				}
			}
		}
	}
	logger.Debugf("Downloaded %d artefacts in %s", count, time.Since(start))
	return nil
}
