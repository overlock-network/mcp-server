package schema

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// CreateProvidersToolInputSchema creates the JSON schema for the get-providers tool input
// This schema matches the QueryListProviderRequest from the Overlock API
func CreateProvidersToolInputSchema() *jsonschema.Schema {
	zero := 0.0
	thousand := 1000.0

	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"creator": {
				Type:        "string",
				Description: "Filter providers by creator address (optional)",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of providers to return (default: 100, max: 1000)",
				Minimum:     &zero,
				Maximum:     &thousand,
			},
			"offset": {
				Type:        "integer",
				Description: "Number of providers to skip for pagination (default: 0)",
				Minimum:     &zero,
			},
		},
		AdditionalProperties: &jsonschema.Schema{},
	}
}
