package dsn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"

	mysqlauthproxy "github.com/block/ftl-mysql-auth-proxy"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
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

// MySQLDSN returns a MySQLDSN string for connecting to the a MySQL database.
func MySQLDSN(dbName string, options ...Option) string {
	opts := &dsnOptions{port: 13306, host: "127.0.0.1", user: "root", password: "secret"}
	for _, opt := range options {
		opt(opts)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowNativePasswords=True", opts.user, opts.password, opts.host, opts.port, dbName)
}

func ResolvePostgresDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		return c.DSN, nil
	case *schema.AWSIAMAuthDatabaseConnector:
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", fmt.Errorf("configuration error: %w", err)
		}

		region, err := parseRegionFromEndpoint(c.Endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse region from endpoint: %w", err)
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, c.Endpoint, region, c.Username, cfg.Credentials)
		if err != nil {
			return "", fmt.Errorf("failed to create authentication token: %w", err)
		}
		host, port, err := net.SplitHostPort(c.Endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to split host and port: %w", err)
		}
		return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", host, port, c.Database, c.Username, authenticationToken), nil
	default:
		return "", fmt.Errorf("unexpected database connector type: %T", connector)
	}
}

func parseRegionFromEndpoint(endpoint string) (string, error) {
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to split host and port: %w", err)
	}
	host = strings.TrimSuffix(host, ".rds.amazonaws.com")
	parts := strings.Split(host, ".")
	return parts[len(parts)-1], nil
}

func ResolveMySQLConfig(ctx context.Context, connector schema.DatabaseConnector) (*mysqlauthproxy.Config, error) {
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		cfg, err := mysqlauthproxy.ParseDSN(c.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DSN: %w", err)
		}
		return cfg, nil

	case *schema.AWSIAMAuthDatabaseConnector:
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("configuration error: %w", err)
		}

		region, err := parseRegionFromEndpoint(c.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to parse region from endpoint: %w", err)
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, c.Endpoint, region, c.Username, cfg.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to create authentication token: %w", err)
		}

		tls, err := tlsForMySQLIAMAuth(c.Endpoint, region)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}

		mcfg := mysqlauthproxy.NewConfig()
		mcfg.User = c.Username
		mcfg.Passwd = authenticationToken
		mcfg.Net = "tcp"
		mcfg.Addr = c.Endpoint
		mcfg.DBName = c.Database
		mcfg.TLS = tls
		mcfg.AllowCleartextPasswords = true

		return mcfg, nil
	default:
		return nil, fmt.Errorf("unexpected database connector type: %T", connector)
	}
}

// TempMySQLProxy creates a temporary MySQL proxy for the given connector.
// It returns the local address of the proxy and an error if the proxy fails to start.
// When connecting to the proxy, any db name can be used, and it will redirect to the given connector
//
// The proxy will be automatically cleaned up when the context is cancelled.
func TempMySQLProxy(ctx context.Context, connector schema.DatabaseConnector) (host string, port int, err error) {
	logger := log.FromContext(ctx)
	portC := make(chan int)
	proxy := mysqlauthproxy.NewProxy("localhost", 0, func(ctx context.Context) (*mysqlauthproxy.Config, error) {
		cfg, err := ResolveMySQLConfig(ctx, connector)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve MySQL DSN: %w", err)
		}
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
		return "", 0, err
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

func tlsForMySQLIAMAuth(endpoint, region string) (*tls.Config, error) {
	// We need to use AWS CA certs for RDS MySQL connections when using IAM auth.
	// We could also use RDS Proxy here to avoid the need for the CA certs in the future.
	rootCertPool := x509.NewCertPool()
	rdsCert, err := rdsCerts.ReadFile("certs/rds/rds-" + region + "-bundle.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read RDS certificate for region %s: %w", region, err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(rdsCert); !ok {
		return nil, errors.New("failed to append PEM")
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		ServerName: host,
		RootCAs:    rootCertPool,
	}, nil
}
