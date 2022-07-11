package random_test

import (
	"testing"

	"github.com/bejaneps/speedtest/internal/pkg/random"
	"github.com/stretchr/testify/assert"
)

func TestStringWithCharset(t *testing.T) {
	limit := 10
	charset := "123456789"

	randString := random.StringWithCharset(limit, charset)
	assert.True(t, len(randString) == limit, len(randString))

	for _, r := range randString {
		assert.True(t, runeInString(charset, r))
	}
}

func TestString(t *testing.T) {
	limit := 10

	randString := random.String(limit)
	assert.True(t, len(randString) == limit, len(randString))

	for _, r := range randString {
		assert.True(t, runeInString("aBcD123.!#", r))
	}
}

func runeInString(hay string, needle rune) bool {
	for _, r := range hay {
		if r == needle {
			return true
		}
	}

	return false
}
