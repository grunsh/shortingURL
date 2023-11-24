package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParams(t *testing.T) {
	assert.Equal(t, 4, len(Prms))
	assert.Equal(t, "a", Prms["servAddr"].param.name)
	assert.Equal(t, "localhost:8080", Prms["servAddr"].param.defValue)
	assert.Equal(t, "b", Prms["baseUrl"].param.name)
	assert.Equal(t, "http://localhost:8080/", Prms["baseUrl"].param.defValue)
}
