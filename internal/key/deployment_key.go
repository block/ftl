package key

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"errors"
	"strings"
)

type Deployment = KeyType[DeploymentPayload, *DeploymentPayload]

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*Deployment)(nil)

func NewDeploymentKey(module string) Deployment { return newKey[DeploymentPayload](module) }
func ParseDeploymentKey(key string) (Deployment, error) {
	return parseKey[DeploymentPayload](key)
}

type DeploymentPayload struct {
	Module string
}

var _ KeyPayload = (*DeploymentPayload)(nil)

func (d *DeploymentPayload) Kind() string   { return "dpl" }
func (d *DeploymentPayload) String() string { return d.Module }
func (d *DeploymentPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module> but got empty string")
	}
	d.Module = strings.Join(parts, "-")
	return nil
}
func (d *DeploymentPayload) RandomBytes() int { return 10 }
