package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProvidersToolInputSchema(t *testing.T) {
	schema := CreateProvidersToolInputSchema()

	require.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.NotNil(t, schema.Properties)
	assert.Len(t, schema.Properties, 3)

	creatorProp := schema.Properties["creator"]
	require.NotNil(t, creatorProp)
	assert.Equal(t, "string", creatorProp.Type)
	assert.Equal(t, "Filter providers by creator address (optional)", creatorProp.Description)

	limitProp := schema.Properties["limit"]
	require.NotNil(t, limitProp)
	assert.Equal(t, "integer", limitProp.Type)
	assert.Equal(t, "Maximum number of providers to return (default: 100, max: 1000)", limitProp.Description)
	require.NotNil(t, limitProp.Minimum)
	assert.Equal(t, 0.0, *limitProp.Minimum)
	require.NotNil(t, limitProp.Maximum)
	assert.Equal(t, 1000.0, *limitProp.Maximum)

	offsetProp := schema.Properties["offset"]
	require.NotNil(t, offsetProp)
	assert.Equal(t, "integer", offsetProp.Type)
	assert.Equal(t, "Number of providers to skip for pagination (default: 0)", offsetProp.Description)
	require.NotNil(t, offsetProp.Minimum)
	assert.Equal(t, 0.0, *offsetProp.Minimum)
	assert.Nil(t, offsetProp.Maximum)

	assert.NotNil(t, schema.AdditionalProperties)
}
