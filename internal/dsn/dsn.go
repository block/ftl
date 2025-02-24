package dsn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"

	mysqlauthproxy "github.com/block/ftl-mysql-auth-proxy"
	"github.com/block/ftl/common/schema"
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

		// if the file exists, skip the download
		if _, err := os.Stat("/tmp/rds-bundle.pem"); os.IsNotExist(err) {
			log.Printf("Downloading CA cert for region: %s", region)
			// download the ca cert
			caCertURL := fmt.Sprintf("https://truststore.pki.rds.amazonaws.com/%s/%s-bundle.pem", region, region)
			caCertFile, err := os.Create("/tmp/rds-bundle.pem")
			if err != nil {
				return nil, fmt.Errorf("failed to create ca cert file: %w", err)
			}
			defer caCertFile.Close()
			resp, err := http.Get(caCertURL)
			if err != nil {
				return nil, fmt.Errorf("failed to get ca cert: %w", err)
			}
			defer resp.Body.Close()
			_, err = io.Copy(caCertFile, resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to copy ca cert: %w", err)
			}
		}

		log.Printf("Parsing CA cert for region: %s", region)
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile("/tmp/rds-bundle.pem")
		if err != nil {
			log.Fatal(err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Fatal("Failed to append PEM.")
		}
		host, _, err := net.SplitHostPort(c.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to split host and port: %w", err)
		}

		mcfg := mysqlauthproxy.NewConfig()
		mcfg.User = c.Username
		mcfg.Passwd = authenticationToken
		mcfg.Net = "tcp"
		mcfg.Addr = c.Endpoint
		mcfg.DBName = c.Database
		mcfg.TLS = &tls.Config{
			ServerName: host,
			RootCAs:    rootCertPool,
		}
		mcfg.AllowCleartextPasswords = true

		return mcfg, nil
	default:
		return nil, fmt.Errorf("unexpected database connector type: %T", connector)
	}
}
