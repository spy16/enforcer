package main

import (
	"context"
	"log"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/httpapi"
	"github.com/spy16/enforcer/rule"
	"github.com/spy16/enforcer/stores/inmem"
)

func main() {
	store := &inmem.Store{}
	enforcerAPI := &enforcer.API{
		Store:  store,
		Engine: rule.New(),
	}

	if err := httpapi.Serve(":8080", enforcerAPI, getActor); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func getActor(_ context.Context, actorID string) (*enforcer.Actor, error) {
	return &enforcer.Actor{
		ID: actorID,
		Attribs: map[string]interface{}{
			"is_blacklisted": false,
			"segments":       []string{"foo", "gbm-foo"},
		},
	}, nil
}
