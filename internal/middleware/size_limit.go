package middleware

import (
	"net/http"
	"strconv"
)

func SizeLimit(sizeLimit int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if contentLength := r.Header.Get("Content-Length"); contentLength != "" {
				cl, err := strconv.ParseInt(contentLength, 10, 64)
				if err == nil {
					if cl > sizeLimit {
						rw, ok := w.(*responseWriter)
						if ok && !rw.written {
							w.WriteHeader(http.StatusRequestEntityTooLarge)
						}
						return
					}
				}
			}

			r.Body = http.MaxBytesReader(w, r.Body, sizeLimit)
			next.ServeHTTP(w, r)
		})
	}
}
