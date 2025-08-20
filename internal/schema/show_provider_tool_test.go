package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProviderToolInputSchema(t *testing.T) {
	schema := CreateProviderToolInputSchema()

	require.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	
	// Check that id property exists
	assert.Contains(t, schema.Properties, "id")
	
	idSchema := schema.Properties["id"]
	assert.Equal(t, "integer", idSchema.Type)
	assert.Equal(t, "Provider ID to retrieve detailed information for (required)", idSchema.Description)
	assert.Equal(t, 1.0, *idSchema.Minimum)

	// Check required fields
	assert.Contains(t, schema.Required, "id")
	assert.Len(t, schema.Required, 1)
	
	// Check additional properties
	assert.NotNil(t, schema.AdditionalProperties)
}