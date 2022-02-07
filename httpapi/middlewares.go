package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		start := time.Now()

		cWr := &captureWriter{ResponseWriter: wr}
		next.ServeHTTP(cWr, req)

		log.Info().
			Str("remote_addr", req.RemoteAddr).
			Str("request_id", middleware.GetReqID(req.Context())).
			Dur("latency", time.Since(start)).
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("status", cWr.status).
			Msg("request handled")
	})
}

type captureWriter struct {
	http.ResponseWriter

	status int
}

func (cw *captureWriter) WriteHeader(status int) {
	cw.ResponseWriter.WriteHeader(status)
	cw.status = status
}
