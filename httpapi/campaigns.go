package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
)

func getCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := strings.TrimSpace(chi.URLParam(req, "id"))

		c, err := api.GetCampaign(req.Context(), campID)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusOK, c)
	}
}

func listCampaigns(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		p := req.URL.Query()
		q := enforcer.Query{
			Include:    cleanSplit(p.Get("include"), ","),
			SearchIn:   cleanSplit(p.Get("search_in"), ","),
			HavingTags: cleanSplit(p.Get("tags"), ","),
			OnlyActive: p.Get("only_active") == "true",
		}

		camps, err := api.ListCampaigns(req.Context(), q)
		if err != nil {
			writeErr(wr, req, err)
			return
		}
		if camps == nil {
			camps = []enforcer.Campaign{}
		}

		writeOut(wr, req, http.StatusOK, camps)
	}
}

func createCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		var c enforcer.Campaign
		if err := json.NewDecoder(req.Body).Decode(&c); err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		created, err := api.CreateCampaign(req.Context(), c)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusCreated, created)
	}
}

func updateCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := strings.TrimSpace(chi.URLParam(req, "id"))

		var upd enforcer.Updates
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			writeErr(wr, req, enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err))
			return
		}

		c, err := api.UpdateCampaign(req.Context(), campID, upd)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusOK, c)
	}
}

func deleteCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID := strings.TrimSpace(chi.URLParam(req, "id"))

		err := api.DeleteCampaign(req.Context(), campID)
		if err != nil {
			writeErr(wr, req, err)
			return
		}

		writeOut(wr, req, http.StatusNoContent)
	}
}

func cleanSplit(s, sep string) []string {
	var res []string
	for _, item := range strings.Split(s, sep) {
		item = strings.TrimSpace(item)
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}
