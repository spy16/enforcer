package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/spy16/enforcer"
)

type genMap map[string]interface{}

func writeErr(wr http.ResponseWriter, req *http.Request, err error) {
	switch {
	case errors.Is(err, enforcer.ErrNotFound):
		writeOut(wr, req, http.StatusNotFound, err)

	case errors.Is(err, enforcer.ErrInvalid):
		writeOut(wr, req, http.StatusBadRequest, err)

	case errors.Is(err, enforcer.ErrIneligible):
		writeOut(wr, req, http.StatusBadRequest, err)

	case errors.Is(err, enforcer.ErrConflict):
		writeOut(wr, req, http.StatusConflict, err)

	default:
		writeOut(wr, req, http.StatusInternalServerError,
			enforcer.ErrInternal.WithCausef(err.Error()))
	}
}

func writeOut(wr http.ResponseWriter, req *http.Request, status int, v ...interface{}) {
	wr.Header().Set("Content-Type", "application/json; charset=utf-8")
	wr.WriteHeader(status)

	if len(v) > 0 {
		if err := json.NewEncoder(wr).Encode(v[0]); err != nil {
			log.Printf("failed to write to response-writer: %v", err)
		}
	}
}

func serveGraceful(ctx context.Context, gracePeriod time.Duration, addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		log.Info().AnErr("reason", ctx.Err()).Msg("server shutting down")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("graceful shutdown failed")
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
