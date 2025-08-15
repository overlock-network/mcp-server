package schema

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// CreateEnvironmentToolInputSchema creates the JSON schema for the show-environment tool input
// This schema matches the QueryShowEnvironmentRequest from the Overlock API
func CreateEnvironmentToolInputSchema() *jsonschema.Schema {
	one := 1.0
	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "integer",
				Description: "Environment ID to retrieve detailed information for (required)",
				Minimum:     &one,
			},
		},
		Required:             []string{"id"},
		AdditionalProperties: &jsonschema.Schema{},
	}
}
