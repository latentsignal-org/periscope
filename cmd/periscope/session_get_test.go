package main

import (
	"bytes"
	"testing"

	"go.kenn.io/agentsview/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintSessionDetailShowsSecretLeak(t *testing.T) {
	d := &service.SessionDetail{}
	d.ID = "s1"
	d.SecretLeakCount = 3
	var buf bytes.Buffer
	require.NoError(t, printSessionDetailHuman(&buf, d))
	out := buf.String()
	assert.Contains(t, out, "Secrets")
	assert.Contains(t, out, "3")
}

func TestPrintSessionDetailHidesZeroSecretLeak(t *testing.T) {
	d := &service.SessionDetail{}
	d.ID = "s1"
	d.SecretLeakCount = 0
	var buf bytes.Buffer
	require.NoError(t, printSessionDetailHuman(&buf, d))
	assert.NotContains(t, buf.String(), "Secrets",
		"clean session should not show a Secrets line")
}
