package middleware

import (
	"net"
	"net/http"

	"github.com/lingrino/go-fault"
)

var _ fault.Injector = (*ConnectionResetInjector)(nil)

// Injects a connection error by closing the connection immediately while also
// simulating a TCP RST via setting SO_LINGER to 0.
type ConnectionResetInjector struct{}

func (i *ConnectionResetInjector) Handler(_ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)

		if !ok {
			http.Error(w, "connection hijacking not supported", http.StatusInternalServerError)
			return
		}

		conn, _, err := hijacker.Hijack()

		if err != nil {
			http.Error(w, "failed to hijack connection", http.StatusInternalServerError)
			return
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetLinger(0) // Best effort RST on close
		}

		conn.Close()
	})
}
