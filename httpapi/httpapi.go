package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/spy16/enforcer"
)

type getActor func(ctx context.Context, actorID string) (*enforcer.Actor, error)

type campaignsAPI interface {
	GetCampaign(ctx context.Context, name string) (*enforcer.Campaign, error)
	ListCampaigns(ctx context.Context, q enforcer.Query) ([]enforcer.Campaign, error)
	CreateCampaign(ctx context.Context, c enforcer.Campaign) (*enforcer.Campaign, error)
	UpdateCampaign(ctx context.Context, name string, updates enforcer.Updates) (*enforcer.Campaign, error)
	DeleteCampaign(ctx context.Context, name string) error
}

type enrolmentsAPI interface {
	GetEnrolment(ctx context.Context, campName string, ac enforcer.Actor) (*enforcer.Enrolment, error)
	ListAllEnrolments(ctx context.Context, ac enforcer.Actor, q enforcer.Query) ([]enforcer.Enrolment, error)
	Enrol(ctx context.Context, campaignName string, act enforcer.Actor) (*enforcer.Enrolment, bool, error)
	Ingest(ctx context.Context, completeMulti bool, ac enforcer.Actor, act enforcer.Action) ([]enforcer.Enrolment, error)
}

// Serve starts an REST api server on given bind address.
func Serve(addr string, enforcerAPI *enforcer.API, getActor getActor) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/ping", pingHandler())
	r.Route("/v1/campaigns", func(r chi.Router) {
		r.Get("/", listCampaigns(enforcerAPI))
		r.Post("/", createCampaign(enforcerAPI))
		r.Get("/{id}", getCampaign(enforcerAPI))
		r.Put("/{id}", updateCampaign(enforcerAPI))
		r.Delete("/{id}", deleteCampaign(enforcerAPI))
	})

	r.Route("/v1/actors/{actor_id}", func(r chi.Router) {
		r.Get("/enrolments/{campaign_name}", getEnrolment(enforcerAPI, getActor))
		r.Get("/enrolments", listEnrolments(enforcerAPI, getActor))
		r.Post("/enrol", enrol(enforcerAPI, getActor))
		r.Post("/ingest", ingest(enforcerAPI, getActor))
	})

	return http.ListenAndServe(addr, r)
}

func pingHandler() http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		writeOut(wr, req, http.StatusOK, genMap{"status": "ok"})
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

type genMap map[string]interface{}
