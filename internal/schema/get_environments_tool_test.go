package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEnvironmentsToolInputSchema(t *testing.T) {
	schema := CreateEnvironmentsToolInputSchema()

	require.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)

	// Check required properties exist
	assert.Contains(t, schema.Properties, "creator")
	assert.Contains(t, schema.Properties, "limit")
	assert.Contains(t, schema.Properties, "offset")

	// Check creator property
	creatorProp := schema.Properties["creator"]
	assert.Equal(t, "string", creatorProp.Type)
	assert.Equal(t, "Filter environments by creator address (optional)", creatorProp.Description)

	// Check limit property
	limitProp := schema.Properties["limit"]
	assert.Equal(t, "integer", limitProp.Type)
	assert.Equal(t, "Maximum number of environments to return (default: 100, max: 1000)", limitProp.Description)
	assert.NotNil(t, limitProp.Minimum)
	assert.Equal(t, 0.0, *limitProp.Minimum)
	assert.NotNil(t, limitProp.Maximum)
	assert.Equal(t, 1000.0, *limitProp.Maximum)

	// Check offset property
	offsetProp := schema.Properties["offset"]
	assert.Equal(t, "integer", offsetProp.Type)
	assert.Equal(t, "Number of environments to skip for pagination (default: 0)", offsetProp.Description)
	assert.NotNil(t, offsetProp.Minimum)
	assert.Equal(t, 0.0, *offsetProp.Minimum)

	// Check no required fields (all optional)
	assert.Empty(t, schema.Required)

	// Check additional properties
	assert.NotNil(t, schema.AdditionalProperties)
}