package main

import (
	"testing"
	"time"

	plugin "github.com/CodeClarityCE/plugin-codeql/src"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	out := plugin.Start("../../js-sbom/tests/npmv1", time.Now())

	// Assert the expected values
	assert.NotNil(t, out)
}

func BenchmarkCreate(b *testing.B) {
}
