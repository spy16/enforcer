package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/spy16/enforcer"
)

// Serve starts an REST api server on given bind address.
func Serve(ctx context.Context, addr string, enforcerAPI *enforcer.API, getActor getActor) error {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		requestLogger,
		middleware.Recoverer,
	)

	r.Get("/ping", pingHandler())
	r.Route("/v1/campaigns", func(r chi.Router) {
		r.Get("/", listCampaigns(enforcerAPI))
		r.Post("/", createCampaign(enforcerAPI))
		r.Get("/{id}", getCampaign(enforcerAPI))
		r.Put("/{id}", updateCampaign(enforcerAPI))
		r.Delete("/{id}", deleteCampaign(enforcerAPI))
	})

	r.Route("/v1/actors/{actor_id}", func(r chi.Router) {
		r.Get("/enrolments/{campaign_id}", getEnrolment(enforcerAPI, getActor))
		r.Get("/enrolments", listEnrolments(enforcerAPI, getActor))
		r.Post("/enrol", enrol(enforcerAPI, getActor))
		r.Post("/ingest", ingest(enforcerAPI, getActor))
	})

	return serveGraceful(ctx, 10*time.Second, addr, r)
}

type getActor func(ctx context.Context, actorID string) (*enforcer.Actor, error)

type campaignsAPI interface {
	GetCampaign(ctx context.Context, id string) (*enforcer.Campaign, error)
	ListCampaigns(ctx context.Context, q enforcer.Query) ([]enforcer.Campaign, error)
	CreateCampaign(ctx context.Context, c enforcer.Campaign) (*enforcer.Campaign, error)
	UpdateCampaign(ctx context.Context, id string, updates enforcer.Updates) (*enforcer.Campaign, error)
	DeleteCampaign(ctx context.Context, id string) error
}

type enrolmentsAPI interface {
	GetEnrolment(ctx context.Context, campaignID string, ac enforcer.Actor) (*enforcer.Enrolment, error)
	ListAllEnrolments(ctx context.Context, ac enforcer.Actor, q enforcer.Query) ([]enforcer.Enrolment, error)
	Enrol(ctx context.Context, campaignID string, act enforcer.Actor) (*enforcer.Enrolment, bool, error)
	Ingest(ctx context.Context, completeMulti bool, ac enforcer.Actor, act enforcer.Action) ([]enforcer.IngestResult, error)
}

func pingHandler() http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		writeOut(wr, req, http.StatusOK, genMap{"status": "ok"})
	}
}
