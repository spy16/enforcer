package enrolment

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/campaign"
)

type Service struct {
	Store     Store
	Campaigns *campaign.Service
}

func (en *Service) GetEnrolment(ctx context.Context, actor Actor, campaignID int) (*Enrolment, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	} else if campaignID <= 0 {
		return nil, enforcer.ErrInvalid.WithMsgf("campaign id must be positive integer, not %d", campaignID)
	}

	enr, err := en.Store.GetEnrolment(ctx, actor.ID, campaignID)
	if !errors.Is(err, enforcer.ErrNotFound) {
		return enr, err
	}

	// [active enrolment not found. figure out if eligible and return]
	camp, err := en.Campaigns.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	return en.prepEnrolment(ctx, actor, *camp)
}

func (en *Service) ListEnrolments(ctx context.Context, actor Actor, q Query) ([]Enrolment, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	}

	needOnlyExisting := len(q.Status) > 0 && !contains(q.Status, StatusEligible)
	existing, err := en.Store.ListEnrolments(ctx, actor.ID, q.Status)
	if err != nil || needOnlyExisting {
		return existing, err
	}

	q.Campaigns.Include = collectCampaignIDs(existing)
	applicableCampaigns, err := en.Campaigns.ListCampaigns(ctx, q.Campaigns)
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
			if errors.Is(err, enforcer.ErrIneligible) {
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
		return nil, enforcer.ErrConflict.WithMsgf("already enrolled")
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

type Event map[string]interface{}

func (en *Service) Ingest(ctx context.Context, actor Actor, query Query, event Event) ([]Enrolment, error) {
	applicableEnrolments, err := en.ListEnrolments(ctx, actor, query)
	if err != nil {
		return nil, err
	}

	var res []Enrolment
	for _, enr := range applicableEnrolments {
		fmt.Println(enr)
	}
	return res, nil
}

func (en *Service) prepEnrolment(ctx context.Context, actor Actor, camp campaign.Campaign) (*Enrolment, error) {
	if camp.MaxEnrolments > 0 && camp.CurEnrolments >= camp.MaxEnrolments {
		return nil, enforcer.ErrIneligible.WithMsgf("already at maximum enrolments")
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

func preparePotentialList(enrolments []Enrolment, campaigns []campaign.Campaign) []potentialEnrolment {
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
	Campaign campaign.Campaign
}

func contains(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}