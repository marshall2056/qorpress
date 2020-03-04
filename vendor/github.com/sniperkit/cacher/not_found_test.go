package httpcache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundLoginShouldReturnError(t *testing.T) {
	nf := &notFound{}
	token, err := nf.New(context.Background())
	assert.NotNil(t, err)
	assert.Equal(t, "Service not found", err.Error())
	assert.Equal(t, "", token)
}
