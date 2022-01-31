package campaign

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/spy16/enforcer"
)

func TestCampaign_IsActive(t *testing.T) {
	t.Parallel()

	now := time.Now()

	table := []struct {
		title    string
		campaign Campaign
		checkAt  time.Time
		want     bool
	}{
		{
			title: "Disabled",
			campaign: Campaign{
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
			},
			checkAt: now,
			want:    false,
		},
		{
			title: "InPast",
			campaign: Campaign{
				Enabled: true,
				StartAt: now.AddDate(-2, 0, 0),
				EndAt:   now.AddDate(-1, 0, 0),
			},
			checkAt: now,
			want:    false,
		},
		{
			title: "InFuture",
			campaign: Campaign{
				Enabled: true,
				StartAt: now.AddDate(1, 0, 0),
				EndAt:   now.AddDate(2, 0, 0),
			},
			checkAt: now,
			want:    false,
		},
		{
			title: "Enabled",
			campaign: Campaign{
				Enabled: true,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
			},
			checkAt: now,
			want:    true,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got := tt.campaign.IsActive(tt.checkAt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCampaign_HasAllTags(t *testing.T) {
	sample := Campaign{
		Scopes: []string{"foo:bar", "type:test", "country:US"},
	}

	assert.True(t, sample.HasScope([]string{"foo:bar", "country:US"}))
	assert.False(t, sample.HasScope([]string{"foo:bar", "product:ride_service"}))
}

func TestCampaign_Validate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	table := []struct {
		title    string
		campaign Campaign
		wantErr  error
	}{
		{
			title: "NoRuleAtAll",
			campaign: Campaign{
				Name:    "foo",
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Scopes:  []string{"", "foo:bar"},
			},
			wantErr: enforcer.ErrInvalid,
		},
		{
			title: "EmptyStep",
			campaign: Campaign{
				Name:    "foo",
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Spec: Spec{
					Steps: []string{"       "},
				},
			},
			wantErr: enforcer.ErrInvalid,
		},
		{
			title: "InvalidDeadline",
			campaign: Campaign{
				Name:    "foo",
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Spec: Spec{
					Eligibility: "not user.blocked",
					Deadline:    -10,
				},
			},
			wantErr: enforcer.ErrInvalid,
		},
		{
			title: "InvalidPriority",
			campaign: Campaign{
				Name:    "foo",
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Spec: Spec{
					Eligibility: "not user.blocked",
					Priority:    100000,
				},
			},
			wantErr: enforcer.ErrInvalid,
		},
		{
			title: "Valid",
			campaign: Campaign{
				Name:    "foo",
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Spec: Spec{
					Eligibility: "not user.blocked",
					Deadline:    10,
					Priority:    50,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got := tt.campaign.Validate(true)
			if tt.wantErr == nil {
				assert.NoError(t, got)
			} else {
				assert.Truef(t, errors.Is(got, tt.wantErr),
					"wanted '%v', got '%v'", tt.wantErr, got)
			}
		})
	}
}

func TestCampaign_merge(t *testing.T) {
	// TODO: add tests once merge() is finished.
}
