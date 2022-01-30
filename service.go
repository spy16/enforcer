package enforcer

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

type Service struct {
	Store Store
}

// GetCampaign returns campaign with given ID. Returns ErrNotFound if not found.
func (en *Service) GetCampaign(ctx context.Context, campaignID int) (*Campaign, error) {
	if campaignID <= 0 {
		return nil, ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}

	res, err := en.Store.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListCampaigns returns a list of campaigns matching the given search query.
func (en *Service) ListCampaigns(ctx context.Context, q Query) ([]Campaign, error) {
	res, err := en.Store.ListCampaigns(ctx, q)
	if err != nil {
		return nil, err
	}
	return q.filterCampaigns(res), nil
}

// CreateCampaign validates and inserts a new campaign into the storage.
// Campaign ID is assigned automatically and the stored version of the
// campaign is returned.
func (en *Service) CreateCampaign(ctx context.Context, camp Campaign) (*Campaign, error) {
	if err := camp.Validate(); err != nil {
		return nil, err
	}

	id, err := en.Store.CreateCampaign(ctx, camp)
	if err != nil {
		return nil, err
	}
	camp.ID = id

	return &camp, nil
}

// UpdateCampaign merges the given partial campaign object with the existing
// campaign and stores. The updated version is returned. Some fields may not
// undergo update based on current usage status.
func (en *Service) UpdateCampaign(ctx context.Context, partial Campaign) (*Campaign, error) {
	if partial.ID <= 0 {
		return nil, ErrInvalid.WithMsgf("invalid id: %d", partial.ID)
	}

	updateFn := func(ctx context.Context, actual *Campaign) error {
		return actual.merge(partial)
	}

	return en.Store.UpdateCampaign(ctx, partial.ID, updateFn)
}

// DeleteCampaign deletes a campaign by the identifier.
func (en *Service) DeleteCampaign(ctx context.Context, campaignID int) error {
	if campaignID <= 0 {
		return ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}
	return en.Store.DeleteCampaign(ctx, campaignID)
}

func (en *Service) GetEnrolment(ctx context.Context, actor Actor, campaignID int) (*Enrolment, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	} else if campaignID <= 0 {
		return nil, ErrInvalid.WithMsgf("campaign id must be positive integer, not %d", campaignID)
	}

	enr, err := en.Store.GetEnrolment(ctx, actor.ID, campaignID)
	if !errors.Is(err, ErrNotFound) {
		return enr, err
	}

	// [active enrolment not found. figure out if eligible and return]
	camp, err := en.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	return en.prepEnrolment(ctx, actor, *camp)
}

func (en *Service) ListEnrolments(ctx context.Context, actor Actor, status []string, q Query) ([]Enrolment, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	}

	needOnlyExisting := len(status) > 0 && !contains(status, StatusEligible)
	existing, err := en.Store.ListEnrolments(ctx, actor.ID, status)
	if err != nil || needOnlyExisting {
		return existing, err
	}

	q.Include = collectCampaignIDs(existing)
	applicableCampaigns, err := en.ListCampaigns(ctx, q)
	if err != nil {
		return nil, err
	}

	potential := preparePotentialList(existing, applicableCampaigns)
	sort.Slice(potential, func(i, j int) bool {
		return potential[i].Weight > potential[j].Weight
	})

	res := append([]Enrolment{}, existing...)
	for _, pe := range potential {
		if pe.Existing != nil {
			continue // already included in the list.
		}

		enr, err := en.prepEnrolment(ctx, actor, pe.Campaign)
		if err != nil {
			if errors.Is(err, ErrIneligible) {
				continue // not eligible for this.
			}
			return nil, err
		}
		res = append(res, *enr)
	}

	return res, nil
}

func (en *Service) Enrol(ctx context.Context, actor Actor, campaignID int) (*Enrolment, error) {
	enr, err := en.GetEnrolment(ctx, actor, campaignID)
	if err != nil {
		return nil, err
	} else if enr.Status != StatusEligible {
		return nil, ErrConflict.WithMsgf("already enrolled")
	}

	enr.Status = StatusActive
	enr.StartedAt = time.Now()
	enr.EndsAt = enr.Campaign.EndAt
	if enr.Campaign.Deadline > 0 {
		deadlineDur := time.Duration(enr.Campaign.Deadline*24) * time.Hour
		enr.EndsAt = enr.StartedAt.Add(deadlineDur)
	}
	if err := en.Store.CreateEnrolment(ctx, *enr); err != nil {
		return nil, err
	}
	return enr, nil
}

func (en *Service) Ingest(ctx context.Context, actor Actor, query Query, event Event) ([]Enrolment, error) {
	applicableEnrolments, err := en.ListEnrolments(ctx, actor, nil, query)
	if err != nil {
		return nil, err
	}

	var res []Enrolment
	for _, enr := range applicableEnrolments {
		fmt.Println(enr)
	}
	return res, nil
}

func (en *Service) prepEnrolment(ctx context.Context, actor Actor, camp Campaign) (*Enrolment, error) {
	if camp.MaxEnrolments > 0 && camp.CurEnrolments >= camp.MaxEnrolments {
		return nil, ErrIneligible.WithMsgf("already at maximum enrolments")
	}

	if camp.Eligibility != "" {
		// TODO: execute eligibility rule here.
	}

	return &Enrolment{
		Status:         StatusEligible,
		ActorID:        actor.ID,
		CampaignID:     camp.ID,
		RemainingSteps: len(camp.Steps),
		Campaign:       &camp,
	}, nil
}

func preparePotentialList(enrolments []Enrolment, campaigns []Campaign) []potentialEnrolment {
	enrIdx := map[int]int{}
	for i, e := range enrolments {
		enrIdx[e.CampaignID] = i
	}

	var res []potentialEnrolment
	for _, c := range campaigns {
		pe := potentialEnrolment{Campaign: c}
		if enrID, found := enrIdx[c.ID]; found {
			pe.Existing = &enrolments[enrID]
		}
		// TODO: compute weight for this.
		res = append(res, pe)
	}
	return res
}

func collectCampaignIDs(enrolments []Enrolment) []int {
	set := map[int]struct{}{}
	res := make([]int, len(enrolments), len(enrolments))
	for i, e := range enrolments {
		if _, found := set[e.CampaignID]; !found {
			res[i] = e.CampaignID
			set[e.CampaignID] = struct{}{}
		}
	}
	return res
}

type potentialEnrolment struct {
	Weight   int
	Existing *Enrolment
	Campaign Campaign
}

func contains(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}
