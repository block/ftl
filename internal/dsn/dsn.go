package dsn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	"github.com/alecthomas/errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	yaml "sigs.k8s.io/yaml/goyaml.v2"

	mysqlauthproxy "github.com/block/ftl-mysql-auth-proxy"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/pgproxy"
)

type dsnOptions struct {
	host     string
	port     int
	user     string
	password string
}

type Option func(*dsnOptions)

func Port(port int) Option {
	return func(o *dsnOptions) {
		o.port = port
	}
}

func Host(host string) Option {
	return func(o *dsnOptions) {
		o.host = host
	}
}

func Username(user string) Option {
	return func(o *dsnOptions) {
		o.user = user
	}
}

func Password(password string) Option {
	return func(o *dsnOptions) {
		o.password = password
	}
}

// PostgresDSN returns a PostgresDSN string for connecting to a PG database.
func PostgresDSN(dbName string, options ...Option) string {
	opts := &dsnOptions{port: 15432, host: "127.0.0.1", user: "postgres", password: "secret"}
	for _, opt := range options {
		opt(opts)
	}
	return fmt.Sprintf("postgres://%s:%d/%s?sslmode=disable&user=%s&password=%s", opts.host, opts.port, dbName, opts.user, opts.password)
}

// MySQLDSN returns a MySQLDSN string for connecting to a MySQL database.
func MySQLDSN(dbName string, options ...Option) string {
	opts := &dsnOptions{port: 13306, host: "127.0.0.1", user: "root", password: "secret"}
	for _, opt := range options {
		opt(opts)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowNativePasswords=True", opts.user, opts.password, opts.host, opts.port, dbName)
}

func ResolvePostgresDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	logger := log.FromContext(ctx)
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		logger.Debugf("Resolving Postgres DSN DSNDatabaseConnector")
		return c.DSN, nil
	case *schema.AWSIAMAuthDatabaseConnector:
		logger.Debugf("Resolving Postgres DSN AWSIAMAuthDatabaseConnector")
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", errors.Wrap(err, "configuration error")
		}

		region, err := parseRegionFromEndpoint(c.Endpoint)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse region from endpoint")
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, c.Endpoint, region, c.Username, cfg.Credentials)
		if err != nil {
			return "", errors.Wrap(err, "failed to create authentication token")
		}
		host, port, err := net.SplitHostPort(c.Endpoint)
		if err != nil {
			return "", errors.Wrap(err, "failed to split host and port")
		}
		return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", host, port, c.Database, c.Username, authenticationToken), nil
	case *schema.YAMLFileCredentialsConnector:
		logger.Debugf("Resolving Postgres DSN YAMLFileCredentialsConnector %s", c.Path)
		return resolveYAMLFileCredentials(ctx, c)
	default:
		return "", errors.Errorf("unexpected database connector type: %T", connector)
	}
}

func resolveYAMLFileCredentials(ctx context.Context, connector *schema.YAMLFileCredentialsConnector) (string, error) {
	bytes, err := os.ReadFile(connector.Path)
	if err != nil {
		return "", errors.Wrap(err, "failed to read DB Credentials file")
	}

	dsn, err := parseDSNFromYAML(ctx, string(bytes), connector.DSNTemplate)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse DSN from YAML")
	}

	return dsn, nil
}

func convertKeysToStrings(m map[any]any) map[string]any {
	converted := make(map[string]any, len(m))
	for k, v := range m {
		ks, ok := k.(string)
		if ok {
			converted[ks] = v
			if m, ok := v.(map[any]any); ok {
				converted[ks] = convertKeysToStrings(m)
			}
		}
	}
	return converted
}

func parseDSNFromYAML(ctx context.Context, yml string, tmplStr string) (string, error) {
	raw := map[any]any{}
	err := yaml.Unmarshal([]byte(yml), &raw)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal YAML")
	}
	cfg := convertKeysToStrings(raw)

	var errResult error
	return os.Expand(tmplStr, func(key string) string {
		builder := gval.Full(jsonpath.Language())
		path, err := builder.NewEvaluable(key)
		if err != nil {
			errResult = err
			return ""
		}
		r, err := path(ctx, cfg)
		if err != nil {
			errResult = err
			return ""
		}
		return fmt.Sprintf("%s", r)
	}), errResult
}

func parseRegionFromEndpoint(endpoint string) (string, error) {
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host and port")
	}
	if !strings.HasSuffix(host, ".rds.amazonaws.com") {
		return "", errors.Errorf("endpoint %s is not an RDS endpoint", endpoint)
	}
	host = strings.TrimSuffix(host, ".rds.amazonaws.com")
	parts := strings.Split(host, ".")
	return parts[len(parts)-1], nil
}

func MySQLDBName(connector schema.DatabaseConnector) (string, error) {
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		return c.Database, nil
	case *schema.AWSIAMAuthDatabaseConnector:
		return c.Database, nil
	default:
		return "", errors.Errorf("unexpected database connector type: %T", connector)
	}
}

func PostgresDBName(connector schema.DatabaseConnector) (string, error) {
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		return c.Database, nil
	case *schema.AWSIAMAuthDatabaseConnector:
		return c.Database, nil
	default:
		return "", errors.Errorf("unexpected database connector type: %T", connector)
	}
}

func ResolveMySQLConfig(ctx context.Context, connector schema.DatabaseConnector) (*mysqlauthproxy.Config, error) {
	logger := log.FromContext(ctx)

	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		logger.Debugf("Resolving MySQL config DSNDatabaseConnector")
		cfg, err := mysqlauthproxy.ParseDSN(c.DSN)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse DSN")
		}
		err = updateTLSForMySQL(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update TLS config")
		}

		return cfg, nil

	case *schema.AWSIAMAuthDatabaseConnector:
		logger.Debugf("Resolving MySQL config AWSIAMAuthDatabaseConnector")
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "configuration error")
		}

		region, err := parseRegionFromEndpoint(c.Endpoint)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse region from endpoint")
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, c.Endpoint, region, c.Username, cfg.Credentials)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create authentication token")
		}

		tls, err := tlsForMySQLRDS(c.Endpoint, region)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create TLS config")
		}

		mcfg := mysqlauthproxy.NewConfig()
		mcfg.User = c.Username
		mcfg.Passwd = authenticationToken
		mcfg.Net = "tcp"
		mcfg.Addr = c.Endpoint
		mcfg.DBName = c.Database
		mcfg.TLS = tls
		mcfg.AllowCleartextPasswords = true

		// TODO: the mysql proxy overwrites the client connection requests with this.
		// We should use the config from the client to the proxy to avoid this.
		mcfg.MultiStatements = true

		return mcfg, nil
	case *schema.YAMLFileCredentialsConnector:
		dsn, err := resolveYAMLFileCredentials(ctx, c)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve YAML file credentials")
		}
		cfg, err := mysqlauthproxy.ParseDSN(dsn)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse DSN")
		}
		err = updateTLSForMySQL(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update TLS config")
		}

		return cfg, nil
	default:
		return nil, errors.Errorf("unexpected database connector type: %T", connector)
	}
}

func updateTLSForMySQL(cfg *mysqlauthproxy.Config) error {
	if cfg.TLS != nil {
		return nil
	}

	region, err := parseRegionFromEndpoint(cfg.Addr)
	if err != nil {
		// not an aws dsn
		return nil
	}

	tls, err := tlsForMySQLRDS(cfg.Addr, region)
	if err != nil {
		return err
	}
	cfg.TLS = tls
	return nil
}

// ConnectorMySQLProxy creates a MySQL proxy for the given connector.
// It returns the local address of the proxy and an error if the proxy fails to start.
// The databse name of the connection to this proxy needs to match the database name in the connector.
//
// The proxy will be automatically closed when the context is cancelled.
func ConnectorMySQLProxy(ctx context.Context, connector schema.DatabaseConnector) (host string, port int, err error) {
	logger := log.FromContext(ctx)
	portC := make(chan int)
	proxy := mysqlauthproxy.NewProxy("localhost", 0, func(ctx context.Context) (*mysqlauthproxy.Config, error) {
		cfg, err := ResolveMySQLConfig(ctx, connector)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve MySQL DSN")
		}
		// TODO: the mysql proxy overwrites the client connection requests with this.
		// We should use the config from the client to the proxy to avoid this.
		cfg.MultiStatements = true
		return cfg, nil
	}, &mysqlLogger{logger: logger}, portC)

	errC := make(chan error)
	go func() {
		err := proxy.ListenAndServe(ctx)
		if err != nil {
			logger.Errorf(err, "failed to listen and serve")
			errC <- err
		}
	}()

	select {
	case err := <-errC:
		return "", 0, errors.WithStack(err)
	case port := <-portC:
		logger.Debugf("Started a temporary mysql proxy on port %d", port)
		return "127.0.0.1", port, nil
	}
}

type mysqlLogger struct {
	logger *log.Logger
}

func (m *mysqlLogger) Print(v ...any) {
	for _, s := range v {
		m.logger.Infof("mysqlproxy: %v", s)
	}
}

//go:embed certs/rds
var rdsCerts embed.FS

func tlsForMySQLRDS(endpoint, region string) (*tls.Config, error) {
	// We need to use AWS CA certs for RDS MySQL connections when using IAM auth.
	// We could also use RDS Proxy here to avoid the need for the CA certs in the future.
	rootCertPool := x509.NewCertPool()
	rdsCert, err := rdsCerts.ReadFile("certs/rds/rds-" + region + "-bundle.pem")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read RDS certificate for region %s", region)
	}
	if ok := rootCertPool.AppendCertsFromPEM(rdsCert); !ok {
		return nil, errors.WithStack(errors.New("failed to append PEM"))
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to split host and port")
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		ServerName: host,
		RootCAs:    rootCertPool,
	}, nil
}

// ConnectorPGProxy creates a Postgres proxy for the given connector.
// It returns the local address of the proxy and an error if the proxy fails to start.
// The proxy redirects all database connections to the databse in the given connector.
//
// The proxy will be automatically closed when the context is cancelled.
func ConnectorPGProxy(ctx context.Context, connector schema.DatabaseConnector) (host string, port int, err error) {
	logger := log.FromContext(ctx)
	proxy := pgproxy.New("127.0.0.1:0", func(ctx context.Context, params map[string]string) (string, error) {
		dsn, err := ResolvePostgresDSN(ctx, connector)
		if err != nil {
			return "", errors.Wrap(err, "failed to resolve Postgres DSN")
		}
		return dsn, nil
	})

	started := make(chan pgproxy.Started)
	errC := make(chan error)
	go func() {
		err := proxy.Start(ctx, started)
		if err != nil {
			logger.Errorf(err, "failed to start proxy")
			errC <- err
		}
	}()

	select {
	case err := <-errC:
		return "", 0, errors.WithStack(err)
	case started := <-started:
		logger.Debugf("Started a temporary postgres proxy on port %d", started.Address.Port)
		return "127.0.0.1", started.Address.Port, nil
	}
}
