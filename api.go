package enforcer

import (
	"context"
	"errors"
	"strings"
	"time"
)

// API provides functions for managing campaigns.
type API struct {
	Store  Store
	Engine ruleEngine
}

type ruleEngine interface {
	Exec(_ context.Context, rule string, data interface{}) (bool, error)
}

// GetCampaign returns campaign with given ID. Returns ErrNotFound if not found.
func (api *API) GetCampaign(ctx context.Context, id string) (*Campaign, error) {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return nil, ErrInvalid.
			WithMsgf("invalid campaign id '%s'", id).
			WithCausef("must match '%s'", idPattern)
	}
	return api.Store.GetCampaign(ctx, id)
}

// ListCampaigns returns a list of campaigns matching the given search query.
func (api *API) ListCampaigns(ctx context.Context, q Query) ([]Campaign, error) {
	res, err := api.Store.ListCampaigns(ctx, q)
	if err != nil {
		return nil, err
	}
	return q.filterCampaigns(res), nil
}

// CreateCampaign validates and inserts a new campaign into the storage. Campaign ID is
// assigned automatically and the stored version of the campaign is returned.
func (api *API) CreateCampaign(ctx context.Context, camp Campaign) (*Campaign, error) {
	if err := camp.Validate(); err != nil {
		return nil, err
	}

	if err := api.Store.CreateCampaign(ctx, camp); err != nil {
		return nil, err
	}
	return &camp, nil
}

// UpdateCampaign merges the given partial campaign object with the existing campaign and
// stores. The updated version is returned. Some fields may not undergo update
// based on current usage status.
func (api *API) UpdateCampaign(ctx context.Context, id string, updates Updates) (*Campaign, error) {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return nil, ErrInvalid.
			WithMsgf("invalid campaign id '%s'", id).
			WithCausef("must match '%s'", idPattern)
	}

	updateFn := func(ctx context.Context, actual *Campaign) error {
		if err := actual.apply(updates); err != nil {
			return err
		}
		actual.UpdatedAt = time.Now()
		return nil
	}

	return api.Store.UpdateCampaign(ctx, id, updateFn)
}

// DeleteCampaign deletes a campaign by the identifier.
func (api *API) DeleteCampaign(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return ErrInvalid.
			WithMsgf("invalid campaign id '%s'", id).
			WithCausef("must match '%s'", idPattern)
	}

	return api.Store.DeleteCampaign(ctx, id)
}

// GetEnrolment returns an enrolment for campaign and an actor. If actor is not
// already enrolled into the campaign and is eligible, a virtual enrolment with
// status StatusEligible is returned.
func (api *API) GetEnrolment(ctx context.Context, id string, ac Actor) (*Enrolment, error) {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return nil, ErrInvalid.
			WithMsgf("invalid campaign id '%s'", id).
			WithCausef("must match '%s'", idPattern)
	}

	enr, err := api.Store.GetEnrolment(ctx, ac.ID, id)
	if !errors.Is(err, ErrNotFound) {
		enr.setStatus()
		return enr, err
	}

	camp, err := api.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}

	return api.prepEnrolment(ctx, *camp, ac)
}

// ListExistingEnrolments returns a list of existing enrolments in one of given statuses.
// The returned list will not include eligible enrolments.
func (api *API) ListExistingEnrolments(ctx context.Context, actorID string, status []string) ([]Enrolment, error) {
	existing, err := api.Store.ListEnrolments(ctx, actorID)
	if err != nil {
		return nil, err
	}
	return filterByStatus(existing, status), nil
}

// ListAllEnrolments returns all enrolments including existing and eligible. Eligible
// are computed based on the campaign query and actor data provided.
func (api *API) ListAllEnrolments(ctx context.Context, ac Actor, campQ Query) ([]Enrolment, error) {
	existing, err := api.ListExistingEnrolments(ctx, ac.ID, nil)
	if err != nil {
		return nil, err
	}

	campQ.OnlyActive = true
	campQ.Include = collectCampaignIDs(existing)
	camps, err := api.ListCampaigns(ctx, campQ)
	if err != nil {
		return nil, err
	}

	var res []Enrolment
	alreadyEnrolled := map[string]struct{}{}
	for _, enrolment := range existing {
		enrolment.setStatus()
		alreadyEnrolled[enrolment.CampaignID] = struct{}{}
		res = append(res, enrolment)
	}

	for _, camp := range camps {
		if _, exists := alreadyEnrolled[camp.ID]; exists {
			continue
		}

		enr, err := api.prepEnrolment(ctx, camp, ac)
		if err != nil {
			if errors.Is(err, ErrIneligible) {
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
func (api *API) Enrol(ctx context.Context, campaignID string, ac Actor) (*Enrolment, bool, error) {
	campaignID = strings.TrimSpace(campaignID)
	if !idPattern.MatchString(campaignID) {
		return nil, false, ErrInvalid.
			WithMsgf("invalid campaign id '%s'", campaignID).
			WithCausef("must match '%s'", idPattern)
	}

	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignID)
	if !errors.Is(err, ErrNotFound) {
		enr.setStatus()
		return enr, false, err
	}

	camp, err := api.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, false, err
	}

	newEnr, err := api.prepEnrolment(ctx, *camp, ac)
	if err != nil {
		return nil, false, err
	}

	newEnr.StartedAt = time.Now()
	newEnr.EndsAt = camp.EndAt
	if camp.Deadline > 0 {
		// relative end_date due to deadline (in days)
		newEnr.EndsAt = newEnr.StartedAt.AddDate(0, 0, camp.Deadline)
	}
	newEnr.setStatus()

	return newEnr, true, api.Store.UpsertEnrolment(ctx, *newEnr)
}

// Ingest processes the action within current enrolments and returns the list of
// enrolments that progressed. If completeMulti is false, only one enrolment will
// be progressed.
func (api *API) Ingest(ctx context.Context, completeMulti bool, ac Actor, act Action) ([]IngestResult, error) {
	if err := act.Validate(); err != nil {
		return nil, err
	}

	applicable, err := api.ListExistingEnrolments(ctx, ac.ID, []string{StatusActive})
	if err != nil {
		return nil, err
	}
	api.sortApplicable(applicable)

	var res []IngestResult
	var isAffected bool
	var completionErr error
	for _, enr := range applicable {
		isAffected, completionErr = api.applyCompletion(ctx, ac, act, &enr)
		if completionErr != nil {
			break
		} else if isAffected {
			if completionErr = api.Store.UpsertEnrolment(ctx, enr); completionErr != nil {
				break
			}
			res = append(res, IngestResult{
				StepID:     enr.CompletedSteps[len(enr.CompletedSteps)-1].StepID,
				ActionID:   act.ID,
				CampaignID: enr.CampaignID,
			})
			if !completeMulti {
				break
			}
		}
	}
	return res, completionErr
}

func (api *API) sortApplicable(applicable []Enrolment) {
	// TODO: sort based on priority, end_date etc.
}

func (api *API) prepEnrolment(ctx context.Context, camp Campaign, ac Actor) (*Enrolment, error) {
	if err := api.checkEligibility(ctx, camp, ac); err != nil {
		return nil, err
	}

	return &Enrolment{
		Status:     StatusEligible,
		ActorID:    ac.ID,
		CampaignID: camp.ID,
		TotalSteps: len(camp.Steps),
	}, nil
}

func (api *API) checkEligibility(ctx context.Context, camp Campaign, ac Actor) error {
	if camp.Eligibility == "" {
		return nil
	}

	isPass, err := api.Engine.Exec(ctx, camp.Eligibility, ruleExecEnv(ac, nil))
	if err != nil {
		return err
	} else if !isPass {
		return ErrIneligible
	}
	return nil
}

func (api *API) applyCompletion(ctx context.Context, ac Actor, act Action, enr *Enrolment) (bool, error) {
	camp, err := api.GetCampaign(ctx, enr.CampaignID)
	if err != nil {
		return false, err
	}
	env := ruleExecEnv(ac, &act)

	if camp.IsUnordered {
		done := map[int]struct{}{}
		for _, step := range enr.CompletedSteps {
			done[step.StepID] = struct{}{}
		}
		for i, step := range camp.Steps {
			if _, alreadyDone := done[i]; alreadyDone {
				continue
			}

			pass, err := api.Engine.Exec(ctx, step, env)
			if err != nil {
				return false, err
			} else if pass {
				enr.CompletedSteps = append(enr.CompletedSteps, StepResult{
					StepID:   i,
					DoneAt:   act.Time,
					ActionID: act.ID,
				})
				enr.TotalSteps = len(camp.Steps)
				return true, nil
			}
		}

		return false, nil
	}

	nextStepID := len(enr.CompletedSteps)
	if nextStepID >= len(camp.Steps) {
		return false, ErrInternal.WithMsgf("campaign has lesser steps than enrolment")
	}

	pass, err := api.Engine.Exec(ctx, camp.Steps[nextStepID], env)
	if err != nil || !pass {
		return false, err
	}

	enr.CompletedSteps = append(enr.CompletedSteps, StepResult{
		StepID:   nextStepID,
		DoneAt:   act.Time,
		ActionID: act.ID,
	})
	enr.TotalSteps = len(camp.Steps)
	return true, nil
}

type IngestResult struct {
	StepID     int    `json:"step_id"`
	ActionID   string `json:"action_id"`
	CampaignID string `json:"campaign_id"`
}
