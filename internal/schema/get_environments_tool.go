package schema

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// CreateEnvironmentsToolInputSchema creates the JSON schema for the get-environments tool input
// This schema matches the QueryListEnvironmentRequest from the Overlock API
func CreateEnvironmentsToolInputSchema() *jsonschema.Schema {
	zero := 0.0
	thousand := 1000.0

	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"creator": {
				Type:        "string",
				Description: "Filter environments by creator address (optional)",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of environments to return (default: 100, max: 1000)",
				Minimum:     &zero,
				Maximum:     &thousand,
			},
			"offset": {
				Type:        "integer",
				Description: "Number of environments to skip for pagination (default: 0)",
				Minimum:     &zero,
			},
		},
		AdditionalProperties: &jsonschema.Schema{},
	}
}