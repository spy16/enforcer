package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/actor"
	"github.com/spy16/enforcer/core/campaign"
)

func getEnrolment(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		campName := chi.URLParam(req, "campaign_name")
		enr, err := api.Get(req.Context(), campName, *ac)
		if err != nil {
			if errors.Is(err, enforcer.ErrNotFound) {
				writeOut(wr, req, http.StatusNotFound,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else {
				writeOut(wr, req, http.StatusInternalServerError,
					enforcer.ErrInternal.WithCausef(err.Error()))
			}
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
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		p := req.URL.Query()
		q := campaign.Query{
			OnlyActive:  p.Get("only_active") == "true",
			Include:     cleanSplit(p.Get("include"), ","),
			SearchIn:    cleanSplit(p.Get("search_in"), ","),
			HavingScope: cleanSplit(p.Get("scope"), ","),
		}

		enrolmentList, err := api.ListAll(req.Context(), *ac, q)
		if err != nil {
			writeOut(wr, req, http.StatusInternalServerError,
				enforcer.ErrInternal.WithCausef(err.Error()))
			return
		}

		writeOut(wr, req, http.StatusOK, enrolmentList)
	}
}

func enrol(api enrolmentsAPI, getActor getActor) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		var body struct {
			CampaignName string `json:"campaign_name"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}

		enr, isNew, err := api.Enrol(req.Context(), body.CampaignName, *ac)
		if err != nil {
			if errors.Is(err, enforcer.ErrNotFound) {
				writeOut(wr, req, http.StatusNotFound,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else if errors.Is(err, enforcer.ErrIneligible) {
				writeOut(wr, req, http.StatusBadRequest,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else {
				writeOut(wr, req, http.StatusInternalServerError,
					enforcer.ErrInternal.WithCausef(err.Error()))
			}
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
			Multi  bool         `json:"multi"`
			Action actor.Action `json:"action"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		actorID := chi.URLParam(req, "actor_id")
		ac, err := getActor(req.Context(), actorID)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("actor '%s' not found", actorID))
			return
		}
		body.Action.Actor = *ac

		enr, err := api.Ingest(req.Context(), body.Multi, body.Action)
		if err != nil {
			if errors.Is(err, enforcer.ErrNotFound) {
				writeOut(wr, req, http.StatusNotFound,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else if errors.Is(err, enforcer.ErrIneligible) {
				writeOut(wr, req, http.StatusBadRequest,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else {
				writeOut(wr, req, http.StatusInternalServerError,
					enforcer.ErrInternal.WithCausef(err.Error()))
			}
			return
		}

		writeOut(wr, req, http.StatusOK, enr)
	}
}
