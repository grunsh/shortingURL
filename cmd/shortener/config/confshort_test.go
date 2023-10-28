package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParams(t *testing.T) {
	assert.Equal(t, 3, len(Prms))
	assert.Equal(t, "a", Prms[0].param.name)
	assert.Equal(t, "localhost:8080", Prms[0].param.defValue)
	assert.Equal(t, "b", Prms[1].param.name)
	assert.Equal(t, "http://localhost:8080/", Prms[1].param.defValue)
}
