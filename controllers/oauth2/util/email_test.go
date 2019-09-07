package util_test

import (
	"github.com/tribehq/platform/controllers/oauth2/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	assert.False(t, util.ValidateEmail("test@user"))
	assert.True(t, util.ValidateEmail("test@user.com"))
}
