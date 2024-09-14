package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.Equal(t, "500, err", Internal("500, %v", "err").Error())
	assert.Equal(t, "400, err", BadRequest("400, %v", "err").Error())
	assert.Equal(t, "401, err", Unauthorized("401, %v", "err").Error())
	assert.Equal(t, "403, err", Forbidden("403, %v", "err").Error())
	assert.Equal(t, "404, err", NotFound("404, %v", "err").Error())
	assert.Equal(t, "xxx, err", New("xxx, %v", "err").Error())
}

func TestStatus(t *testing.T) {
	assert.Equal(t, 401, Unauthorized("401, %v", "err").(*Error).HTTP())
}
