package enforcer_test

import (
	goerrors "errors"
	"testing"

	"github.com/spy16/enforcer"
	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		err   enforcer.Error
		want  string
	}{
		{
			title: "WithoutCause",
			err:   enforcer.ErrInvalid,
			want:  "bad_request: request is not valid",
		},
		{
			title: "WithCause",
			err:   enforcer.ErrInvalid.WithCausef("foo"),
			want:  "bad_request: foo",
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		err   enforcer.Error
		other error
		want  bool
	}{
		{
			title: "NonTimerErr",
			err:   enforcer.ErrInternal,
			other: goerrors.New("foo"),
			want:  false,
		},
		{
			title: "TimerErrWithDifferentCode",
			err:   enforcer.ErrInternal,
			other: enforcer.ErrInvalid,
			want:  false,
		},
		{
			title: "TimerErrWithSameCodeDiffCause",
			err:   enforcer.ErrInvalid.WithCausef("cause 1"),
			other: enforcer.ErrInvalid.WithCausef("cause 2"),
			want:  true,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got := goerrors.Is(tt.err, tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_WithCausef(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		err   enforcer.Error
		want  enforcer.Error
	}{
		{
			title: "WithCauseString",
			err:   enforcer.ErrInvalid.WithCausef("foo"),
			want: enforcer.Error{
				Code:    "bad_request",
				Message: "Request is not valid",
				Cause:   "foo",
			},
		},
		{
			title: "WithCauseFormatted",
			err:   enforcer.ErrInvalid.WithCausef("hello %s", "world"),
			want: enforcer.Error{
				Code:    "bad_request",
				Message: "Request is not valid",
				Cause:   "hello world",
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err)
		})
	}
}

func TestError_WithMsgf(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		err   enforcer.Error
		want  enforcer.Error
	}{
		{
			title: "WithCauseString",
			err:   enforcer.ErrInvalid.WithMsgf("foo"),
			want: enforcer.Error{
				Code:    "bad_request",
				Message: "foo",
			},
		},
		{
			title: "WithCauseFormatted",
			err:   enforcer.ErrInvalid.WithMsgf("hello %s", "world"),
			want: enforcer.Error{
				Code:    "bad_request",
				Message: "hello world",
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err)
		})
	}
}

func Test_Errorf(t *testing.T) {
	e := enforcer.Errorf("failed: %d", 100)
	assert.Error(t, e)
	assert.EqualError(t, e, "internal_error: failed: 100")
}
