package testutil

import (
	"testing"

	"github.com/manifestival/manifestival"
	"github.com/stretchr/testify/assert"
)

func TestManifests(t *testing.T, source manifestival.Source) manifestival.Manifest {
	m, err := manifestival.ManifestFrom(source)
	assert.NoError(t, err)
	return m
}
