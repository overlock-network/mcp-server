package schema

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// CreateProviderToolInputSchema creates the JSON schema for the show-provider tool input
// This schema matches the QueryShowProviderRequest from the Overlock API
func CreateProviderToolInputSchema() *jsonschema.Schema {
	one := 1.0
	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "integer",
				Description: "Provider ID to retrieve detailed information for (required)",
				Minimum:     &one,
			},
		},
		Required:             []string{"id"},
		AdditionalProperties: &jsonschema.Schema{},
	}
}