package rpc

import "github.com/alecthomas/errors"

// ServerStream helps to create mockable streaming RPC services.
//
// [connect.ServerStream] is impossible to mock, but we can work around it by substituting our [ServerStream] interface
// into each method and writing a forwarding object that implements the expected Connect interface.
//
// See the tests in this package for an example.
type ServerStream[T any] interface {
	Send(msg *T) error
}

func NewServerStream[T any]() (chan *T, ServerStream[T]) {
	ch := make(chan *T, 16)
	return ch, &serverStream[T]{ch: ch}
}

type serverStream[T any] struct {
	ch chan *T
}

func (s *serverStream[T]) Send(msg *T) error {
	select {
	case s.ch <- msg:
		return nil
	default:
		return errors.New("stream destination channel is closed")
	}
}
