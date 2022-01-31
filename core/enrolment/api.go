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

type campaignsAPI interface {
	Get(ctx context.Context, name string) (*campaign.Campaign, error)
	List(ctx context.Context, q campaign.Query) ([]campaign.Campaign, error)
}

// Enrol binds the given actor to the campaign. Boolean flag will be set only if a new
// enrolment is created.
func (api *API) Enrol(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, bool, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.setStatus()
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

// Get returns an enrolment for campaign and an actor. If actor is not already enrolled into
// the campaign and is eligible, a virtual enrolment with status StatusEligible is returned.
func (api *API) Get(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.setStatus()
		return enr, err
	}

	camp, err := api.CampaignsAPI.Get(ctx, campaignName)
	if err != nil {
		return nil, err
	}

	return api.prepEnrolment(ctx, *camp, ac)
}

// List returns a list of enrolments with given statuses. If the status includes StatusEligible,
// then all eligible enrolments are returned as well.
func (api *API) List(ctx context.Context, ac actor.Actor, status []string, campQ campaign.Query) ([]Enrolment, error) {
	onlyExisting := !contains(status, StatusEligible)
	existing, err := api.Store.ListEnrolments(ctx, ac.ID)
	if err != nil || onlyExisting {
		for i := range existing {
			existing[i].setStatus()
		}
		return filterByStatus(existing, status), err
	}

	camps, err := api.CampaignsAPI.List(ctx, campQ)
	if err != nil {
		return nil, err
	}

	res := append([]Enrolment{}, existing...)
	for _, camp := range camps {
		enr, err := api.prepEnrolment(ctx, camp, ac)
		if err != nil {
			if errors.Is(err, enforcer.ErrIneligible) {
				continue
			}
			return nil, err
		}
		res = append(res, *enr)
	}
	return res, nil
}

func (api *API) prepEnrolment(ctx context.Context, camp campaign.Campaign, ac actor.Actor) (*Enrolment, error) {
	if err := api.checkEligibility(ctx, camp, ac); err != nil {
		return nil, err
	}

	return &Enrolment{
		Status:         StatusEligible,
		ActorID:        ac.ID,
		CampaignID:     camp.Name,
		RemainingSteps: len(camp.Spec.Steps),
	}, nil
}

func (api *API) checkEligibility(ctx context.Context, camp campaign.Campaign, ac actor.Actor) error {
	// TODO: check eligibility.
	return enforcer.ErrIneligible
}

func contains(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}

func filterByStatus(arr []Enrolment, status []string) []Enrolment {
	var res []Enrolment
	for _, enr := range arr {
		enr.setStatus()
		if contains(status, enr.Status) {
			res = append(res, enr)
		}
	}
	return res
}
