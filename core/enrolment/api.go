package enrolment

import (
	"context"
	"errors"
	"log"
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

// Get returns an enrolment for campaign and an actor. If actor is not already
// enrolled into the campaign and is eligible, a virtual enrolment with status
// StatusEligible is returned.
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

// ListExisting returns a list of existing enrolments in one of given statuses.
// The returned list will not include eligible enrolments.
func (api *API) ListExisting(ctx context.Context, actorID string, status []string) ([]Enrolment, error) {
	existing, err := api.Store.ListEnrolments(ctx, actorID)
	if err != nil {
		return nil, err
	}
	return filterByStatus(existing, status), nil
}

// ListAll returns all enrolments including existing and eligible. Eligible enrolments
// are computed based on the campaign query provided and actor data.
func (api *API) ListAll(ctx context.Context, ac actor.Actor, campQ campaign.Query) ([]Enrolment, error) {
	existing, err := api.ListExisting(ctx, ac.ID, nil)
	if err != nil {
		return nil, err
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

// Enrol binds the given actor to the campaign. Boolean flag will be set only if
// a new enrolment is created.
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

	newEnr, err := api.prepEnrolment(ctx, *camp, ac)
	if err != nil {
		return nil, false, err
	}

	newEnr.StartedAt = time.Now()
	newEnr.EndsAt = camp.EndAt
	if camp.Spec.Deadline > 0 {
		// relative end_date due to deadline (in days)
		newEnr.EndsAt = newEnr.StartedAt.AddDate(0, 0, camp.Spec.Deadline)
	}
	newEnr.setStatus()

	return newEnr, true, api.Store.UpsertEnrolment(ctx, *newEnr)
}

// Ingest processes the action within current enrolments and returns the list of
// enrolments that progressed. If completeMulti is false, only one enrolment will
// be progressed.
func (api *API) Ingest(ctx context.Context, completeMulti bool, act actor.Action) ([]Enrolment, error) {
	applicable, err := api.ListExisting(ctx, act.Actor.ID, []string{StatusActive})
	if err != nil {
		return nil, err
	}

	var res []Enrolment
	var isAffected bool
	var completionErr error
	for _, enr := range applicable {
		isAffected, completionErr = api.applyCompletion(ctx, act, &enr)
		if completionErr != nil {
			break
		} else if isAffected {
			if completionErr = api.Store.UpsertEnrolment(ctx, enr); completionErr != nil {
				break
			}
			res = append(res, enr)
		}
	}
	return res, completionErr
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

func (api *API) applyCompletion(ctx context.Context, act actor.Action, enr *Enrolment) (bool, error) {
	camp, err := api.CampaignsAPI.Get(ctx, enr.CampaignID)
	if err != nil {
		return false, err
	}
	log.Printf("checking completion for '%s'", camp.Name)

	// TODO: figure out step completion procedure.

	return false, nil
}