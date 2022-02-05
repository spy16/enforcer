package enforcer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnrolment_computeStatus(t *testing.T) {
	t.Parallel()

	now := time.Now()

	table := []struct {
		title      string
		enr        Enrolment
		wantStatus string
	}{
		{
			title:      "Eligible",
			enr:        Enrolment{},
			wantStatus: StatusEligible,
		},
		{
			title: "Active",
			enr: Enrolment{
				StartedAt:  now.AddDate(0, 0, -10),
				EndsAt:     now.AddDate(0, 0, 10),
				TotalSteps: 3,
			},
			wantStatus: StatusActive,
		},
		{
			title: "Active",
			enr: Enrolment{
				StartedAt:  now.AddDate(0, 0, -10),
				EndsAt:     now.AddDate(0, 0, -5),
				TotalSteps: 3,
			},
			wantStatus: StatusExpired,
		},
		{
			title: "Completed_AndEnded",
			enr: Enrolment{
				StartedAt:  now.AddDate(0, 0, -10),
				EndsAt:     now.AddDate(0, 0, -5),
				TotalSteps: 1,
				CompletedSteps: []StepResult{
					{ActionID: "step_event_1"},
				},
			},
			wantStatus: StatusCompleted,
		},
		{
			title: "Completed_NotEnded",
			enr: Enrolment{
				StartedAt:  now.AddDate(0, 0, -10),
				EndsAt:     now.AddDate(0, 0, 3),
				TotalSteps: 1,
				CompletedSteps: []StepResult{
					{ActionID: "step_event_1"},
				},
			},
			wantStatus: StatusCompleted,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.enr.setStatus()
			assert.Equal(t, tt.wantStatus, tt.enr.Status)
		})
	}
}

func TestEnrolment_validate(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		sample  Enrolment
		wantErr bool
	}{
		{
			title: "Invalid",
			sample: Enrolment{
				CompletedSteps: []StepResult{
					{
						StepID:   0,
						ActionID: "ORDER/1234",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			err := tt.sample.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
