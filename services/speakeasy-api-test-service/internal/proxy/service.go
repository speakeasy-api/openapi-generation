package proxy

import (
	"bufio"
	"net/http"
	"os"
)

const proxiedHost = "bypassed.com:80"

func getForwardURL() string {
	port := os.Getenv("HTTPBIN_PORT")
	if port == "" {
		port = "35123"
	}
	return "http://localhost:" + port + "/anything"
}

// ProxyHandler handles HTTP CONNECT tunneling. It hijacks the connection,
// reads the tunneled HTTP request, and forwards it to httpbin's /anything
// endpoint which echoes the request back. Only CONNECT requests targeting
// proxiedHost are intercepted; all others pass through.
func ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect || r.RequestURI != proxiedHost {
			next.ServeHTTP(w, r)
			return
		}

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijacking not supported", http.StatusInternalServerError)
			return
		}

		conn, buf, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

		tunnelReq, err := http.ReadRequest(bufio.NewReader(buf))
		if err != nil {
			return
		}
		defer tunnelReq.Body.Close()

		proxyReq, err := http.NewRequest(tunnelReq.Method, getForwardURL(), tunnelReq.Body)
		if err != nil {
			return
		}
		for key, values := range tunnelReq.Header {
			for _, v := range values {
				proxyReq.Header.Add(key, v)
			}
		}

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			buf.WriteString("HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n")
			buf.Flush()
			return
		}
		defer resp.Body.Close()

		resp.Write(conn)
	})
}
