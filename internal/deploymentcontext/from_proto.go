package deploymentcontext

import (
	"fmt"

	errors "github.com/alecthomas/errors"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
)

func DBTypeFromProto(x ftlv1.GetDeploymentContextResponse_DbType) DBType {
	switch x {
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED:
		return DBTypeUnspecified
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_POSTGRES:
		return DBTypePostgres
	case ftlv1.GetDeploymentContextResponse_DB_TYPE_MYSQL:
		return DBTypeMySQL
	default:
		panic(fmt.Sprintf("unknown DB type: %d", x))
	}
}

func FromProto(response *ftlv1.GetDeploymentContextResponse) (DeploymentContext, error) {
	databases := map[string]Database{}
	for name, entry := range response.Databases {
		db, err := NewDatabase(DBTypeFromProto(entry.Type), entry.Dsn)
		if err != nil {
			return DeploymentContext{}, errors.Wrapf(err, "could not create database %q with DSN %q", name, entry.Dsn)
		}
		databases[entry.Name] = db
	}
	return NewBuilder(response.Module).AddConfigs(response.Configs).AddSecrets(response.Secrets).AddDatabases(databases).Build(), nil
}
