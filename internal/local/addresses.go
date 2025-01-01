package local

import (
	"fmt"
	"net"
)

// FreeTCPAddresses returns a list of local tcp addresses that are free to listen to at the time of the call.
func FreeTCPAddresses(count int) ([]*net.TCPAddr, error) {
	addresses := make([]*net.TCPAddr, count)
	for i := range count {
		l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			return nil, fmt.Errorf("failed to allocate port: %w", err)
		}
		defer l.Close()
		addresses[i] = l.Addr().(*net.TCPAddr) //nolint:forcetypeassert
	}
	return addresses, nil
}
