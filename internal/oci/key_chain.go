package oci

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"
	"sync"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/block/ftl/common/log"
	"github.com/google/go-containerregistry/pkg/authn"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type keyChain struct {
	originalContext context.Context
	targetConfig    RegistryConfig
	registries      map[string]*registryAuth
	registryLock    sync.Mutex
}

// Resolve implements authn.Keychain.
func (k *keyChain) Resolve(r authn.Resource) (authn.Authenticator, error) {
	k.registryLock.Lock()
	defer k.registryLock.Unlock()

	logger := log.FromContext(k.originalContext)
	registry := r.String()
	existing := k.registries[registry]
	if existing != nil {
		return existing, nil
	}
	cfg := &registryAuth{}
	k.registries[registry] = cfg
	cfg.auth.Store(&authn.AuthConfig{})

	if registry == k.targetConfig.Registry &&
		k.targetConfig.Username != "" &&
		k.targetConfig.Password != "" {
		// The user has explicitly supplied credentials, lets use them
		cfg.auth.Store(&authn.AuthConfig{
			Username: k.targetConfig.Username,
			Password: k.targetConfig.Password,
		})
		return cfg, nil
	}

	dctx, err := authn.DefaultKeychain.ResolveContext(k.originalContext, r)
	if err == nil {
		// Local docker config takes precidence
		cfg.delegate = dctx
		return cfg, nil
	}

	if isECRRepository(registry) {
		username, password, err := getECRCredentials(k.originalContext)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		logger.Debugf("Using ECR credentials for registry '%s'", registry)
		cfg.auth.Store(&authn.AuthConfig{Username: username, Password: password})
		go func() {
			for {
				select {
				case <-k.originalContext.Done():
					return
				case <-time.After(time.Hour):
					username, password, err := getECRCredentials(k.originalContext)
					if err != nil {
						logger.Errorf(err, "failed to refresh ECR credentials")
					}
					cfg.auth.Store(&authn.AuthConfig{Username: username, Password: password})
				}
			}
		}()
	}
	return cfg, nil
}

func isECRRepository(repo string) bool {
	ecrRegex := regexp.MustCompile(`(?i)^\d{12}\.dkr\.ecr\.[a-z0-9-]+\.amazonaws\.com/`)
	return ecrRegex.MatchString(repo)
}

func getECRCredentials(ctx context.Context) (string, string, error) {
	// Load AWS Config
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to load AWS config")
	}

	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)
	// Get authorization token
	resp, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get authorization token")
	}

	if len(resp.AuthorizationData) == 0 {
		return "", "", errors.Wrap(err, "no authorization data")
	}
	authData := resp.AuthorizationData[0]
	token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to decode auth token")
	}

	splitToken := strings.SplitN(string(token), ":", 2)
	if len(splitToken) != 2 {
		return "", "", errors.Wrap(err, "failed to decode auth token due to invalid format")
	}

	username := splitToken[0]
	password := splitToken[1]
	return username, password, nil
}
