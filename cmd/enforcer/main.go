package main

import (
	"github.com/spy16/enforcer/core/campaign"
	"github.com/spy16/enforcer/core/enrolment"
	"github.com/spy16/enforcer/core/rule"
	"github.com/spy16/enforcer/httpapi"
	"github.com/spy16/enforcer/stores/inmem"
)

func main() {
	store := &inmem.Store{}

	campsAPI := &campaign.API{Store: store}
	enrsAPI := &enrolment.API{
		Store:        store,
		RuleEngine:   rule.New(),
		CampaignsAPI: campsAPI,
	}

	if err := httpapi.Serve(":8080", campsAPI, enrsAPI); err != nil {
		panic(err)
	}
}
