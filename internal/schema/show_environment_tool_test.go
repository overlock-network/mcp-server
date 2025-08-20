package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEnvironmentToolInputSchema(t *testing.T) {
	schema := CreateEnvironmentToolInputSchema()

	require.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.NotNil(t, schema.Properties)
	assert.Len(t, schema.Properties, 1)

	idProp := schema.Properties["id"]
	require.NotNil(t, idProp)
	assert.Equal(t, "integer", idProp.Type)
	assert.Equal(t, "Environment ID to retrieve detailed information for (required)", idProp.Description)
	require.NotNil(t, idProp.Minimum)
	assert.Equal(t, 1.0, *idProp.Minimum)

	assert.Len(t, schema.Required, 1)
	assert.Equal(t, "id", schema.Required[0])
	assert.NotNil(t, schema.AdditionalProperties)
}
