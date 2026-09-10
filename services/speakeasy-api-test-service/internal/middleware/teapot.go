package middleware

import "net/http"

func Teapot(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		format := r.Header.Get("x-teapot")
		if format == "" {
			h.ServeHTTP(w, r)
			return
		}

		if format == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(`{"message": "I'm a teapot"}`))
			return
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte("I'm a teapot\n"))
			return
		}
	})
}
