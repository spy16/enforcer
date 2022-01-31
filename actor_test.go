package enforcer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActor_Validate(t *testing.T) {
	actor := Actor{ID: ""}
	assert.Error(t, actor.Validate())

	actor.ID = "1234"
	assert.NoError(t, actor.Validate())
}
