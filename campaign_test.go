package enforcer

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		Tags: []string{"foo:bar", "type:test", "country:US"},
	}

	assert.True(t, sample.HasAllTags([]string{"foo:bar", "country:US"}))
	assert.False(t, sample.HasAllTags([]string{"foo:bar", "product:ride_service"}))
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
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
			},
			wantErr: ErrInvalid,
		},
		{
			title: "EmptyStep",
			campaign: Campaign{
				Enabled: false,
				StartAt: now.AddDate(0, 0, -3),
				EndAt:   now.AddDate(0, 0, 3),
				Steps:   []string{"               "},
			},
			wantErr: ErrInvalid,
		},
		{
			title: "InvalidDeadline",
			campaign: Campaign{
				Enabled:     false,
				StartAt:     now.AddDate(0, 0, -3),
				EndAt:       now.AddDate(0, 0, 3),
				Eligibility: "not user.blocked",
				Deadline:    -10,
			},
			wantErr: ErrInvalid,
		},
		{
			title: "InvalidPriority",
			campaign: Campaign{
				Enabled:     false,
				StartAt:     now.AddDate(0, 0, -3),
				EndAt:       now.AddDate(0, 0, 3),
				Eligibility: "not user.blocked",
				Priority:    100000,
			},
			wantErr: ErrInvalid,
		},
		{
			title: "Valid",
			campaign: Campaign{
				Enabled:     false,
				StartAt:     now.AddDate(0, 0, -3),
				EndAt:       now.AddDate(0, 0, 3),
				Eligibility: "not user.blocked",
				Deadline:    10,
				Priority:    50,
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got := tt.campaign.Validate()
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
