package key

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"

	errors "github.com/alecthomas/errors"
)

type Deployment = KeyType[DeploymentPayload, *DeploymentPayload]

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*Deployment)(nil)

func NewDeploymentKey(realm, module string) Deployment {
	if realm == "" {
		panic("realm is required")
	}
	if module == "" {
		panic("module is required")
	}

	return newKey[DeploymentPayload](realm, module)
}
func ParseDeploymentKey(key string) (Deployment, error) {
	return errors.WithStack2(parseKey[DeploymentPayload](key))
}

type DeploymentPayload struct {
	Realm  string
	Module string
}

var _ KeyPayload = (*DeploymentPayload)(nil)

func (d *DeploymentPayload) Kind() string   { return "dpl" }
func (d *DeploymentPayload) String() string { return fmt.Sprintf("%s-%s", d.Realm, d.Module) }
func (d *DeploymentPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.Errorf("expected <realm>-<module> but got %s", strings.Join(parts, "-"))
	}
	d.Realm = parts[0]
	d.Module = parts[1]
	return nil
}
func (d *DeploymentPayload) RandomBytes() int { return 10 }
