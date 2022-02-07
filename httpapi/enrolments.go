package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
)

func getEnrolment(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		actorID := chi.URLParam(req, "actor_id")
		campID := chi.URLParam(req, "campaign_id")

		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		enr, err := api.GetEnrolment(req.Context(), campID, *ac)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusOK, enr)
	}
}

func listEnrolments(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		p := req.URL.Query()
		q := enforcer.Query{
			OnlyActive: p.Get("only_active") == "true",
			Include:    cleanSplit(p.Get("include"), ","),
			SearchIn:   cleanSplit(p.Get("search_in"), ","),
			HavingTags: cleanSplit(p.Get("tags"), ","),
		}

		enrolmentList, err := api.ListAllEnrolments(req.Context(), *ac, q)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusOK, enrolmentList)
	}
}

func enrol(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		var body struct {
			CampaignID string `json:"campaign_id"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		enr, isNew, err := api.Enrol(req.Context(), body.CampaignID, *ac)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		if isNew {
			writeOut(wr, req, http.StatusCreated, enr)
		} else {
			writeOut(wr, req, http.StatusOK, enr)
		}
	}
}

func ingest(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		var body struct {
			Multi  bool            `json:"multi"`
			Action enforcer.Action `json:"action"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}
		body.Action.ActorID = ac.ID

		enr, err := api.Ingest(req.Context(), body.Multi, *ac, body.Action)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		if enr == nil {
			enr = []enforcer.Enrolment{}
		}
		writeOut(wr, req, http.StatusOK, enr)
	}
}
