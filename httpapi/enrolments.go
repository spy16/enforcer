package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
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

		campID, err := strconv.ParseInt(chi.URLParam(req, "campaign_id"), 10, 64)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithMsgf("campaign id must be an integer").WithCausef(err.Error()))
			return
		}

		enr, err := api.GetEnrolment(req.Context(), int(campID), *ac)
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
		q := enforcer.Query{
			OnlyActive:  p.Get("only_active") == "true",
			Include:     intArray(cleanSplit(p.Get("include"), ",")),
			SearchIn:    intArray(cleanSplit(p.Get("search_in"), ",")),
			HavingScope: cleanSplit(p.Get("scope"), ","),
		}

		enrolmentList, err := api.ListAllEnrolments(req.Context(), *ac, q)
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
			CampaignID int `json:"campaign_id"`
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

		enr, isNew, err := api.Enrol(req.Context(), body.CampaignID, *ac)
		if err != nil {
			if errors.Is(err, enforcer.ErrNotFound) {
				writeOut(wr, req, http.StatusNotFound,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else if errors.Is(err, enforcer.ErrIneligible) {
				writeOut(wr, req, http.StatusBadRequest,
					enforcer.ErrNotFound.WithCausef(err.Error()))
			} else if errors.Is(err, enforcer.ErrInvalid) {
				writeOut(wr, req, http.StatusBadRequest, err)
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
			Multi  bool            `json:"multi"`
			Action enforcer.Action `json:"action"`
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

		enr, err := api.Ingest(req.Context(), body.Multi, *ac, body.Action)
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

		if enr == nil {
			enr = []enforcer.Enrolment{}
		}
		writeOut(wr, req, http.StatusOK, enr)
	}
}
