package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction_Validate(t *testing.T) {
	act := Action{ID: ""}
	assert.Error(t, act.Validate())

	act.ID = "order-1234"
	assert.Error(t, act.Validate())

	act.Actor = Actor{ID: "1234"}
	assert.NoError(t, act.Validate())
}

func TestAction_String(t *testing.T) {
	actor := Action{ID: "1234"}
	assert.Equal(t, "Action{id='1234'}", actor.String())
}
