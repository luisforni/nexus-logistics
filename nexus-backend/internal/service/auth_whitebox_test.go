package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailDomain_WithAt(t *testing.T) {
	assert.Equal(t, "@example.com", emailDomain("user@example.com"))
}

func TestEmailDomain_NoAt(t *testing.T) {
	assert.Equal(t, "unknown", emailDomain("noemail"))
}

func TestEmailDomain_EmptyString(t *testing.T) {
	assert.Equal(t, "unknown", emailDomain(""))
}
