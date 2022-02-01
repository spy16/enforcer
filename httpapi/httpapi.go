package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/spy16/enforcer/core/actor"
	"github.com/spy16/enforcer/core/campaign"
	"github.com/spy16/enforcer/core/enrolment"
)

type getActor func(ctx context.Context, actorID string) (*actor.Actor, error)

type campaignsAPI interface {
	Get(ctx context.Context, name string) (*campaign.Campaign, error)
	List(ctx context.Context, q campaign.Query) ([]campaign.Campaign, error)
	Create(ctx context.Context, c campaign.Campaign) (*campaign.Campaign, error)
	Update(ctx context.Context, name string, updates campaign.Updates) (*campaign.Campaign, error)
	Delete(ctx context.Context, name string) error
}

type enrolmentsAPI interface {
	Get(ctx context.Context, campName string, ac actor.Actor) (*enrolment.Enrolment, error)
	ListAll(ctx context.Context, ac actor.Actor, q campaign.Query) ([]enrolment.Enrolment, error)
	Enrol(ctx context.Context, campaignName string, ac actor.Actor) (*enrolment.Enrolment, bool, error)
	Ingest(ctx context.Context, completeMulti bool, act actor.Action) ([]enrolment.Enrolment, error)
}

// Serve starts an REST api server on given bind address.
func Serve(addr string, campsAPI campaignsAPI, enrsAPI enrolmentsAPI,
	getActor getActor) error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/ping", pingHandler())
	r.Route("/v1/campaigns", func(r chi.Router) {
		r.Get("/", listCampaigns(campsAPI))
		r.Post("/", createCampaign(campsAPI))
		r.Get("/{id}", getCampaign(campsAPI))
		r.Put("/{id}", updateCampaign(campsAPI))
		r.Delete("/{id}", deleteCampaign(campsAPI))
	})

	r.Route("/v1/actors/{actor_id}", func(r chi.Router) {
		r.Get("/enrolments/{campaign_name}", getEnrolment(enrsAPI, getActor))
		r.Get("/enrolments", listEnrolments(enrsAPI, getActor))
		r.Post("/enrol", enrol(enrsAPI, getActor))
		r.Post("/ingest", ingest(enrsAPI, getActor))
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
