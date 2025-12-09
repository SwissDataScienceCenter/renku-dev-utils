package httputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTChar(t *testing.T) {
	value := "hello"
	res, rem, err := ParseTChar(value)
	assert.NoError(t, err)
	assert.Equal(t, "h", res)
	assert.Equal(t, "ello", rem)

	value = ":ello"
	res, rem, err = ParseTChar(value)
	assert.ErrorContains(t, err, "not a tchar")
}

func TestParseToken(t *testing.T) {
	value := "token"
	res, rem, err := ParseToken(value)
	assert.NoError(t, err)
	assert.Equal(t, "token", res)
	assert.Equal(t, "", rem)

	value = "hello world"
	res, rem, err = ParseToken(value)
	assert.NoError(t, err)
	assert.Equal(t, "hello", res)
	assert.Equal(t, " world", rem)

	value = ":no"
	res, rem, err = ParseToken(value)
	assert.ErrorContains(t, err, "not a token")
}

func TestParseSPPLus(t *testing.T) {
	value := "   "
	res, rem, err := ParseSPPlus(value)
	assert.NoError(t, err)
	assert.Equal(t, "   ", res)
	assert.Equal(t, "", rem)

	value = "  token"
	res, rem, err = ParseSPPlus(value)
	assert.NoError(t, err)
	assert.Equal(t, "  ", res)
	assert.Equal(t, "token", rem)

	value = "token"
	res, rem, err = ParseSPPlus(value)
	assert.ErrorContains(t, err, "not a space char")
}
