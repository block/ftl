package deploymentcontext

import (
	"context"
	"strings"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/encoding"
)

// DatabasesFromSecrets finds DSNs in secrets and creates a map of databases.
//
// Secret keys should be in the format FTL_DSN_<MODULENAME>_<DBNAME>
func DatabasesFromSecrets(ctx context.Context, module string, secrets map[string][]byte, types map[string]DBType) (map[string]Database, error) {
	databases := map[string]Database{}
	for sName, maybeDSN := range secrets {
		if !strings.HasPrefix(sName, "FTL_DSN_") {
			continue
		}
		// FTL_DSN_<MODULE>_<DBNAME>
		parts := strings.Split(sName, "_")
		if len(parts) != 4 {
			return nil, errors.Errorf("invalid DSN secret key %q should have format FTL_DSN_<MODULE>_<DBNAME>", sName)
		}
		moduleName := strings.ToLower(parts[2])
		dbName := strings.ToLower(parts[3])
		if !strings.EqualFold(moduleName, module) {
			continue
		}
		var dsn string
		if err := encoding.Unmarshal(maybeDSN, &dsn); err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal DSN %q", maybeDSN)
		}
		db, err := NewDatabase(types[dbName], dsn)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create database %q with DSN %q", dbName, maybeDSN)
		}
		databases[dbName] = db
	}
	return databases, nil
}
