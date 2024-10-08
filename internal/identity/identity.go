package identity

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/internal/model"
)

type Identity interface {
	String() string
	identity()
}

var _ Identity = Controller{}

type Controller struct{}

func NewController() Controller {
	return Controller{}
}

func (c Controller) String() string {
	return "ca"
}
func (Controller) identity() {}

var _ Identity = Runner{}

// Runner identity
// TODO: Maybe use KeyType[T any, TP keyPayloadConstraint[T]]?
type Runner struct {
	Key    model.RunnerKey
	Module string
}

func NewRunner(key model.RunnerKey, module string) Runner {
	return Runner{
		Key:    key,
		Module: module,
	}
}

func (r Runner) String() string {
	return fmt.Sprintf("%s:%s", r.Key, r.Module)
}

func (Runner) identity() {}

func Parse(s string) (Identity, error) {
	if s == "" {
		return nil, fmt.Errorf("empty identity")
	}
	parts := strings.Split(s, ":")

	if parts[0] == "ca" && len(parts) == 1 {
		return Controller{}, nil
	}

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid identity: %s", s)
	}

	key, err := model.ParseRunnerKey(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	return NewRunner(key, parts[1]), nil
}

// Wallet is held by a node and contains the node's identity, key pair, signer, and certificate.
type Wallet struct {
	Identity           Identity
	KeyPair            KeyPair
	Signer             Signer
	Certificate        optional.Option[Certificate]
	ControllerVerifier optional.Option[Verifier]
}

func NewStore(identity Identity, keyPair KeyPair) (Wallet, error) {
	signer, err := keyPair.Signer()
	if err != nil {
		return Wallet{}, fmt.Errorf("failed to get signer: %w", err)
	}

	return Wallet{
		Identity: identity,
		KeyPair:  keyPair,
		Signer:   signer,
	}, nil
}

func NewStoreNewKeys(identity Identity) (Wallet, error) {
	pair, err := GenerateKeyPair()
	if err != nil {
		return Wallet{}, fmt.Errorf("failed to generate key pair: %w", err)
	}

	signer, err := pair.Signer()
	if err != nil {
		return Wallet{}, fmt.Errorf("failed to get signer: %w", err)
	}

	return Wallet{
		Identity: identity,
		KeyPair:  pair,
		Signer:   signer,
	}, nil
}

func (s Wallet) NewCertificateRequest() (CertificateRequest, error) {
	publicKey, err := s.KeyPair.Public()
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to get public key: %w", err)
	}

	certificateContent := CertificateContent{
		Identity:  s.Identity,
		PublicKey: publicKey,
	}
	encoded, err := proto.Marshal(certificateContent.ToProto())
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to marshal cert content: %w", err)
	}

	signature, err := s.Signer.Sign(encoded)
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to sign cert request: %w", err)
	}

	return CertificateRequest{
		CertificateContent: CertificateContent{
			Identity:  s.Identity,
			PublicKey: publicKey,
		},
		Signature: signature,
	}, nil
}

// SignCertificateRequest is called by the controller to sign a certificate request,
// while verifiying the node's signature.
//
// The caller must ensure that the request is valid and the identity is legit.
func (s *Wallet) SignCertificateRequest(req CertificateRequest) (Certificate, error) {
	encoded, err := proto.Marshal(req.CertificateContent.ToProto())
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to marshal cert content: %w", err)
	}
	// Ensure the given pubkey matches the signature.
	verifier, err := NewVerifier(req.CertificateContent.PublicKey)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create verifier for pubkey:%x %w", req.CertificateContent.PublicKey.Bytes, err)
	}
	if err = verifier.Verify(req.Signature, encoded); err != nil {
		return Certificate{}, fmt.Errorf("failed to verify signature: %w", err)
	}

	// Request is valid, sign it.
	signed, err := s.Signer.Sign(encoded)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create ca signed data for cert: %w", err)
	}
	return Certificate{
		CertificateContent: req.CertificateContent,
		Signature:          signed,
	}, nil
}

func (s *Wallet) SetCertificate(cert Certificate, controllerVerifier Verifier) error {
	if err := cert.Verify(controllerVerifier); err != nil {
		return fmt.Errorf("failed to verify controller certificate: %w", err)
	}

	// Verify the certificate is for us, checking identity and public key.
	if cert.Identity.String() != s.Identity.String() {
		return fmt.Errorf("certificate identity does not match: %s != %s", cert.Identity, s.Identity)
	}
	myPub, err := s.KeyPair.Public()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	if !bytes.Equal(myPub.Bytes, cert.PublicKey.Bytes) {
		return fmt.Errorf("certificate public key does not match")
	}

	s.Certificate = optional.Some(cert)
	s.ControllerVerifier = optional.Some(controllerVerifier)
	return nil
}

func (s *Wallet) CertifiedSign(message []byte) (CertifiedSignedData, error) {
	certificate, ok := s.Certificate.Get()
	if !ok {
		return CertifiedSignedData{}, fmt.Errorf("certificate not set")
	}

	signature, err := s.Signer.Sign(message)
	if err != nil {
		return CertifiedSignedData{}, fmt.Errorf("failed to sign data: %w", err)
	}

	return CertifiedSignedData{
		Certificate: certificate,
		Message:     message,
		Signature:   signature,
	}, nil
}
