package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/spy16/enforcer"
)

func getCampaign(api campaignsAPI) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		campID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithMsgf("campaign id must be an integer").WithCausef(err.Error()))
			return
		}

		c, err := api.GetCampaign(req.Context(), int(campID))
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
		q := enforcer.Query{
			OnlyActive:  p.Get("only_active") == "true",
			Include:     intArray(cleanSplit(p.Get("include"), ",")),
			SearchIn:    intArray(cleanSplit(p.Get("search_in"), ",")),
			HavingScope: cleanSplit(p.Get("scope"), ","),
		}

		camps, err := api.ListCampaigns(req.Context(), q)
		if err != nil {
			writeOut(wr, req, http.StatusInternalServerError,
				enforcer.ErrInternal.WithCausef(err.Error()))
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
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err.Error()))
			return
		}

		created, err := api.CreateCampaign(req.Context(), c)
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
		campID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithMsgf("campaign id must be an integer").WithCausef(err.Error()))
			return
		}

		var upd enforcer.Updates
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithCausef("failed to parse body: %v", err.Error()))
			return
		}

		c, err := api.UpdateCampaign(req.Context(), int(campID), upd)
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
		campID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeOut(wr, req, http.StatusBadRequest,
				enforcer.ErrInvalid.WithMsgf("campaign id must be an integer").WithCausef(err.Error()))
			return
		}

		err = api.DeleteCampaign(req.Context(), int(campID))
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

func intArray(arr []string) []int {
	var res []int
	for _, item := range arr {
		id, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			continue
		}

		res = append(res, int(id))
	}
	return res
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
