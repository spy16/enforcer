package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActor_Validate(t *testing.T) {
	actor := Actor{ID: ""}
	assert.Error(t, actor.Validate())

	actor.ID = "1234"
	assert.Error(t, actor.Validate())

	actor.Type = "user"
	assert.NoError(t, actor.Validate())
}

func TestActor_String(t *testing.T) {
	actor := Actor{ID: "1234", Type: "user"}
	assert.Equal(t, "Actor{1234, type='user'}", actor.String())
}
