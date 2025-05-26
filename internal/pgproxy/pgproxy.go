package pgproxy

import (
	"context"
	"io"
	"net"

	"github.com/alecthomas/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/block/ftl/common/log"
)

type Config struct {
	Listen string `name:"listen" short:"l" help:"Address to listen on." env:"FTL_PROXY_PG_LISTEN" default:"127.0.0.1:5678"`
}

// PgProxy is a configurable proxy for PostgreSQL connections
type PgProxy struct {
	listenAddress      string
	connectionStringFn func(ctx context.Context, params map[string]string) (string, error)
}

// DSNConstructor is a function that constructs a new connection string from parameters of the incoming connection.
//
// parameters are pg connection parameters as described in https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
type DSNConstructor func(ctx context.Context, params map[string]string) (string, error)

// New creates a new PgProxy.
//
// address is the address to listen on for incoming connections.
// connectionFn is a function that constructs a new connection string from parameters of the incoming connection.
func New(listenAddress string, connectionFn DSNConstructor) *PgProxy {
	return &PgProxy{
		listenAddress:      listenAddress,
		connectionStringFn: connectionFn,
	}
}

type Started struct {
	Address *net.TCPAddr
}

// Start the proxy
func (p *PgProxy) Start(ctx context.Context, started chan<- Started) error {
	logger := log.FromContext(ctx)

	logger.Debugf("starting pgproxy on %s", p.listenAddress)
	listener, err := net.Listen("tcp", p.listenAddress)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s", p.listenAddress)
	}
	defer listener.Close()

	if started != nil {
		addr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			panic("failed to get TCP address")
		}
		started <- Started{Address: addr}
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf(err, "failed to accept connection")
			continue
		}
		logger.Debugf("Accepted postgres connection from %s", conn.RemoteAddr())
		go HandleConnection(ctx, conn, p.connectionStringFn)
	}
}

// HandleConnection proxies a single connection.
//
// This should be run as the first thing after accepting a connection.
// It will block until the connection is closed.
func HandleConnection(ctx context.Context, conn net.Conn, connectionFn DSNConstructor) {
	defer conn.Close()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "pgproxy: connection closed"))

	logger := log.FromContext(ctx)
	logger.Debugf("new connection established: %s", conn.RemoteAddr())

	backend, startup, err := connectBackend(ctx, conn)
	if err != nil {
		logger.Errorf(err, "failed to connect backend")
		return
	}
	if backend == nil {
		logger.Infof("client disconnected without startup message: %s", conn.RemoteAddr())
		return
	}
	logger.Tracef("startup message: %+v", startup)
	logger.Tracef("backend connected: %s", conn.RemoteAddr())

	hijacked, err := connectFrontend(ctx, connectionFn, startup)
	if err != nil {
		// try again, in case there was a credential rotation
		logger.Debugf("failed to connect frontend: %s, trying again", err)

		hijacked, err = connectFrontend(ctx, connectionFn, startup)
		if err != nil {
			handleBackendError(ctx, backend, err)
			return
		}
	}
	backend.Send(&pgproto3.AuthenticationOk{})
	logger.Debugf("frontend connected")
	for key, value := range hijacked.ParameterStatuses {
		backend.Send(&pgproto3.ParameterStatus{Name: key, Value: value})
	}

	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if err := backend.Flush(); err != nil {
		logger.Errorf(err, "failed to flush backend authentication ok")
		return
	}

	if err := proxy(ctx, backend, hijacked.Frontend); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Warnf("disconnecting %s due to: %s", conn.RemoteAddr(), err)
		}
		return
	}
	logger.Debugf("terminating connection to %s", conn.RemoteAddr())
}

func handleBackendError(ctx context.Context, backend *pgproto3.Backend, err error) {
	logger := log.FromContext(ctx)
	logger.Errorf(err, "backend error")
	backend.Send(&pgproto3.ErrorResponse{
		Severity: "FATAL",
		Message:  err.Error(),
	})
	if err := backend.Flush(); err != nil {
		logger.Errorf(err, "failed to flush backend error response")
	}
}

// connectBackend establishes a connection according to https://www.postgresql.org/docs/current/protocol-flow.html
func connectBackend(ctx context.Context, conn net.Conn) (*pgproto3.Backend, *pgproto3.StartupMessage, error) {
	logger := log.FromContext(ctx)

	backend := pgproto3.NewBackend(conn, conn)

	for {
		startup, err := backend.ReceiveStartupMessage()
		if errors.Is(err, io.EOF) {
			// some clients just terminate the connection and open a new one if it does not support SSL / GSS encryption
			return nil, nil, nil
		} else if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to receive startup message from %s", conn.RemoteAddr())
		}

		logger.Debugf("received startup message: %T from %s", startup, conn.RemoteAddr())

		switch startup := startup.(type) {
		case *pgproto3.SSLRequest:
			// The client is requesting SSL connection. We don't support it.
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return nil, nil, errors.Wrap(err, "failed to write ssl request response")
			}
		case *pgproto3.CancelRequest:
			// TODO: implement cancel requests
			return backend, nil, nil
		case *pgproto3.StartupMessage:
			return backend, startup, nil
		case *pgproto3.GSSEncRequest:
			// The client is requesting GSS encryption. We don't support it.
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return nil, nil, errors.Wrap(err, "failed to write gss encryption request response")
			}
		default:
			return nil, nil, errors.Errorf("unknown startup message: %T", startup)
		}
	}
}

func connectFrontend(ctx context.Context, connectionFn DSNConstructor, startup *pgproto3.StartupMessage) (*pgconn.HijackedConn, error) {
	dsn, err := connectionFn(ctx, startup.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct dsn")
	}

	conn, err := pgconn.Connect(ctx, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to backend")
	}
	hijacked, err := conn.Hijack()
	if err != nil {
		return nil, errors.Wrap(err, "failed to hijack backend")
	}
	return hijacked, nil
}

func proxy(ctx context.Context, backend *pgproto3.Backend, frontend *pgproto3.Frontend) error {
	logger := log.FromContext(ctx)
	errorsChan := make(chan error, 2)

	go func() {
		for {
			msg, err := backend.Receive()
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				errorsChan <- errors.Wrap(err, "failed to receive backend message")
				return
			}
			logger.Tracef("backend message: %T", msg)
			frontend.Send(msg)
			err = frontend.Flush()
			if err != nil {
				errorsChan <- errors.Wrap(err, "failed to receive backend message")
				return
			}
			if _, ok := msg.(*pgproto3.Terminate); ok {
				return
			}
		}
	}()

	go func() {
		for {
			msg, err := frontend.Receive()
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					// if the frontend just closes the connection, we don't need to log an error
					errorsChan <- nil
					return
				}
				errorsChan <- errors.Wrap(err, "failed to receive frontend message")
				return
			}
			logger.Tracef("frontend message: %T", msg)
			backend.Send(msg)
			err = backend.Flush()
			if err != nil {
				errorsChan <- errors.Wrap(err, "failed to receive backend message")
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context done")
		case err := <-errorsChan:
			return errors.WithStack(err)
		}
	}
}
