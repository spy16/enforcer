package main

import (
	"context"
	"log"

	"github.com/spy16/enforcer/core/actor"
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

	if err := httpapi.Serve(":8080", campsAPI, enrsAPI, getActor); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func getActor(_ context.Context, actorID string) (*actor.Actor, error) {
	return &actor.Actor{
		ID: actorID,
		Attribs: map[string]interface{}{
			"is_blacklisted": false,
			"segments":       []string{"foo", "gbm-foo"},
		},
	}, nil
}
