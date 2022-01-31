package campaign

import (
	"context"
	"strings"

	"github.com/spy16/enforcer"
)

// API provides functions for managing campaigns.
type API struct{ Store Store }

// Get returns campaign with given ID. Returns ErrNotFound if not found.
func (api *API) Get(ctx context.Context, name string) (*Campaign, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, enforcer.ErrInvalid.WithMsgf("name must not be empty")
	}
	return api.Store.GetCampaign(ctx, name, true)
}

// List returns a list of campaigns matching the given search query.
func (api *API) List(ctx context.Context, q Query) ([]Campaign, error) {
	res, err := api.Store.ListCampaigns(ctx, q)
	if err != nil {
		return nil, err
	}
	return q.filterCampaigns(res), nil
}

// Create validates and inserts a new campaign into the storage. Campaign ID is
// assigned automatically and the stored version of the campaign is returned.
func (api *API) Create(ctx context.Context, camp Campaign) (*Campaign, error) {
	if err := camp.Validate(true); err != nil {
		return nil, err
	}

	if err := api.Store.CreateCampaign(ctx, camp); err != nil {
		return nil, err
	}

	return &camp, nil
}

// Update merges the given partial campaign object with the existing campaign and
// stores. The updated version is returned. Some fields may not undergo update
// based on current usage status.
func (api *API) Update(ctx context.Context, name string, newSpec Spec) (*Campaign, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, enforcer.ErrInvalid.WithMsgf("name must not be empty")
	}

	updateFn := func(ctx context.Context, actual *Campaign) error {
		return actual.merge(newSpec)
	}

	return api.Store.UpdateCampaign(ctx, name, updateFn)
}

// Delete deletes a campaign by the identifier.
func (api *API) Delete(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return enforcer.ErrInvalid.WithMsgf("name must not be empty")
	}
	return api.Store.DeleteCampaign(ctx, name)
}
