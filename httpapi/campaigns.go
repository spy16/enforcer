package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/campaign"
)

func getCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := chi.URLParam(req, "id")

		c, err := api.Get(req.Context(), campID)
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

		writeOut(wr, req, http.StatusOK, c)
	}
}

func listCampaigns(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		p := req.URL.Query()
		q := campaign.Query{
			OnlyActive:  p.Get("only_active") == "true",
			Include:     cleanSplit(p.Get("include"), ","),
			SearchIn:    cleanSplit(p.Get("search_in"), ","),
			HavingScope: cleanSplit(p.Get("scope"), ","),
		}

		camps, err := api.List(req.Context(), q)
		if err != nil {
			writeOut(wr, req, http.StatusInternalServerError,
				enforcer.ErrInternal.WithCausef(err.Error()))
			return
		}
		if camps == nil {
			camps = []campaign.Campaign{}
		}

		writeOut(wr, req, http.StatusOK, camps)
	}
}

func createCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		var c campaign.Campaign
		if err := json.NewDecoder(req.Body).Decode(&c); err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err.Error()))
			return
		}

		created, err := api.Create(req.Context(), c)
		if err != nil {
			if errors.Is(err, enforcer.ErrInvalid) {
				writeOut(wr, req, http.StatusBadRequest,
					enforcer.ErrInvalid.WithCausef(err.Error()))
			} else {
				writeOut(wr, req, http.StatusInternalServerError,
					enforcer.ErrInvalid.WithCausef(err.Error()))
			}
			return
		}

		writeOut(wr, req, http.StatusCreated, created)
	}
}

func updateCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := chi.URLParam(req, "id")

		var upd campaign.Updates
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err.Error()))
			return
		}

		c, err := api.Update(req.Context(), campID, upd)
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

		writeOut(wr, req, http.StatusOK, c)
	}
}

func deleteCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := chi.URLParam(req, "id")

		err := api.Delete(req.Context(), campID)
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

		writeOut(wr, req, http.StatusNoContent)
	}
}

func cleanSplit(s string, sep string) []string {
	var res []string
	for _, item := range strings.Split(s, sep) {
		item = strings.TrimSpace(item)
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}
