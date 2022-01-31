package enrolment

import (
	"context"
	"errors"
	"time"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/actor"
	"github.com/spy16/enforcer/core/campaign"
)

// API provides functions for managing enrolments.
type API struct {
	Store        Store
	CampaignsAPI campaignsAPI
}

// Enrol binds the given actor to the campaign. Boolean flag will be set only
// if a new enrolment is created.
func (api *API) Enrol(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, bool, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.computeStatus()
		return enr, false, err
	}

	camp, err := api.CampaignsAPI.Get(ctx, campaignName)
	if err != nil {
		return nil, false, err
	}

	if err := api.checkEligibility(ctx, *camp, ac); err != nil {
		return nil, false, err
	}

	newEnr := &Enrolment{
		Status:         StatusActive,
		ActorID:        ac.ID,
		ActorType:      ac.Type,
		CampaignID:     camp.Name,
		StartedAt:      time.Now(),
		EndsAt:         camp.EndAt,
		RemainingSteps: len(camp.Spec.Steps),
	}

	if camp.Spec.Deadline > 0 {
		// relative end_date due to deadline (in days)
		newEnr.EndsAt = newEnr.StartedAt.AddDate(0, 0, camp.Spec.Deadline)
	}

	return newEnr, true, api.Store.CreateEnrolment(ctx, *newEnr)
}

// Get returns an enrolment for campaign and an actor. If actor is not already
// enrolled into the campaign and is eligible, a virtual enrolment with status
// StatusEligible is returned.
func (api *API) Get(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.computeStatus()
		return enr, err
	}

	camp, err := api.CampaignsAPI.Get(ctx, campaignName)
	if err != nil {
		return nil, err
	}

	if err := api.checkEligibility(ctx, *camp, ac); err != nil {
		return nil, err
	}

	return &Enrolment{
		Status:         StatusEligible,
		ActorID:        ac.ID,
		ActorType:      ac.Type,
		CampaignID:     camp.Name,
		RemainingSteps: len(camp.Spec.Steps),
	}, nil
}

func (api *API) checkEligibility(ctx context.Context, camp campaign.Campaign, ac actor.Actor) error {
	// TODO: check eligibility.
	return enforcer.ErrIneligible
}

type campaignsAPI interface {
	Get(ctx context.Context, name string) (*campaign.Campaign, error)
}
