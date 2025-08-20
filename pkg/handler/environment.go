package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Oudwins/zog"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	overlockv1beta1 "github.com/overlock-network/api/go/node/overlock/crossplane/v1beta1"
	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

// EnvironmentInput represents the input parameters for the show-environment tool
type EnvironmentInput struct {
	Id int `json:"id,omitempty"`
}

// EnvironmentsListInput represents the input parameters for the get-environments tool
type EnvironmentsListInput struct {
	Creator string `json:"creator,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// EnvironmentHandler handles both show-environment and get-environments tool requests
type EnvironmentHandler struct {
	chainClient    overlockv1beta1.QueryClient
	timeout        time.Duration
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewEnvironmentHandler creates a new environment handler
func NewEnvironmentHandler(chainClient overlockv1beta1.QueryClient, timeout time.Duration) *EnvironmentHandler {
	// Configure circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "blockchain-client-env",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 2
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Warn().
				Str("circuit_breaker", name).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("Circuit breaker state changed")
		},
	})

	return &EnvironmentHandler{
		chainClient:    chainClient,
		timeout:        timeout,
		circuitBreaker: cb,
	}
}

// Handle processes the show-environment tool call
func (h *EnvironmentHandler) Handle(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	// Create a logger with request context
	logger := log.With().
		Str("tool", "show-environment").
		Str("request_id", fmt.Sprintf("%p", params)).
		Logger()

	start := time.Now()
	logger.Info().Msg("Processing show-environment request")

	// Apply timeout to the context
	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Define validation schema using Zog
	schema := zog.Struct(zog.Shape{
		"id": zog.Int().GTE(1),
	})

	// Validate input parameters
	var input EnvironmentInput
	arguments := params.Arguments
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	logger.Debug().Interface("arguments", arguments).Msg("Validating input arguments")
	// Parse and validate the arguments
	errs := schema.Parse(arguments, &input)
	if errs != nil {
		logger.Error().Interface("errors", errs).Msg("Input validation failed")
		return nil, fmt.Errorf("validation failed: %v", errs)
	}

	// Check if ID was provided (required field)
	if input.Id == 0 {
		logger.Error().Msg("Environment ID is required")
		return nil, fmt.Errorf("validation failed: environment ID is required")
	}

	logger.Debug().Interface("parsed_input", input).Msg("Input validation successful")

	// Create the request
	req := &overlockv1beta1.QueryShowEnvironmentRequest{
		Id: uint64(input.Id),
	}

	// Log request parameters
	logger.Info().
		Uint64("environment_id", req.Id).
		Msg("Fetching environment from blockchain")

	// Check if chain client is available
	if h.chainClient == nil {
		logger.Error().Msg("gRPC client is not available")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: gRPC connection to blockchain is not available. Please check the connection and try again.",
				},
			},
		}, nil
	}

	// Fetch environment from the chain using circuit breaker protection
	result, err := h.circuitBreaker.Execute(func() (interface{}, error) {
		return h.chainClient.ShowEnvironment(timeoutCtx, req)
	})

	if err != nil {
		logger.Info().Err(err).Msg("Failed to connect to gRPC server - blockchain service unavailable")
		// Check if it's a circuit breaker error
		if err == gobreaker.ErrOpenState {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Blockchain service is currently unavailable (circuit breaker protection active). Please try again later.",
					},
				},
			}, nil
		}
		// Return a user-friendly response instead of propagating the error
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Unable to connect to blockchain service. The service may be temporarily unavailable. Please check your connection and try again later.",
				},
			},
		}, nil
	}

	chainResponse, ok := result.(*overlockv1beta1.QueryShowEnvironmentResponse)
	if !ok || chainResponse == nil {
		logger.Error().Msg("Received invalid response from blockchain service")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Received invalid response from blockchain service. Please try again later.",
				},
			},
		}, nil
	}

	// Check if environment was found
	if chainResponse.Environment == nil {
		logger.Info().Uint64("environment_id", req.Id).Msg("Environment not found")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Environment with ID '%d' not found.", req.Id),
				},
			},
		}, nil
	}

	// Log successful response
	duration := time.Since(start)
	logger.Info().
		Uint64("environment_id", req.Id).
		Dur("duration", duration).
		Msg("Successfully fetched environment")

	// Use the official API response directly
	responseJSON, err := json.MarshalIndent(chainResponse, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal response")
		return nil, fmt.Errorf("failed to marshal environment response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseJSON),
			},
		},
	}, nil
}

// HandleList processes the get-environments tool call
func (h *EnvironmentHandler) HandleList(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	// Create a logger with request context
	logger := log.With().
		Str("tool", "get-environments").
		Str("request_id", fmt.Sprintf("%p", params)).
		Logger()

	start := time.Now()
	logger.Info().Msg("Processing get-environments request")

	// Apply timeout to the context
	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Define validation schema using Zog with default values
	schema := zog.Struct(zog.Shape{
		"creator": zog.String().Default(""),
		"limit":   zog.Int().LTE(1000).Default(100),
		"offset":  zog.Int().Default(0),
	})

	// Validate input parameters (always parse to apply defaults)
	var input EnvironmentsListInput
	arguments := params.Arguments
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	logger.Debug().Interface("arguments", arguments).Msg("Validating input arguments")
	// Parse and validate the arguments
	errs := schema.Parse(arguments, &input)
	if errs != nil {
		logger.Error().Interface("errors", errs).Msg("Input validation failed")
		return nil, fmt.Errorf("validation failed: %v", errs)
	}
	logger.Debug().Interface("parsed_input", input).Msg("Input validation successful")

	// Set default pagination
	req := &overlockv1beta1.QueryListEnvironmentRequest{
		Pagination: &query.PageRequest{
			Limit:  100,
			Offset: 0,
		},
	}

	// Apply validated parameters
	if input.Creator != "" {
		req.Creator = input.Creator
	}
	if input.Limit > 0 {
		req.Pagination.Limit = uint64(input.Limit)
	}
	req.Pagination.Offset = uint64(input.Offset)

	// Log request parameters
	logger.Info().
		Uint64("limit", req.Pagination.Limit).
		Uint64("offset", req.Pagination.Offset).
		Str("creator", req.Creator).
		Msg("Fetching environments from blockchain")

	// Check if chain client is available
	if h.chainClient == nil {
		logger.Error().Msg("gRPC client is not available")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: gRPC connection to blockchain is not available. Please check the connection and try again.",
				},
			},
		}, nil
	}

	// Fetch environments from the chain using circuit breaker protection
	result, err := h.circuitBreaker.Execute(func() (interface{}, error) {
		return h.chainClient.ListEnvironment(timeoutCtx, req)
	})

	if err != nil {
		logger.Info().Err(err).Msg("Failed to connect to gRPC server - blockchain service unavailable")
		// Check if it's a circuit breaker error
		if err == gobreaker.ErrOpenState {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Blockchain service is currently unavailable (circuit breaker protection active). Please try again later.",
					},
				},
			}, nil
		}
		// Return a user-friendly response instead of propagating the error
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Unable to connect to blockchain service. The service may be temporarily unavailable. Please check your connection and try again later.",
				},
			},
		}, nil
	}

	chainResponse, ok := result.(*overlockv1beta1.QueryListEnvironmentResponse)
	if !ok || chainResponse == nil {
		logger.Error().Msg("Received invalid response from blockchain service")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Received invalid response from blockchain service. Please try again later.",
				},
			},
		}, nil
	}

	// Log successful response
	environmentCount := len(chainResponse.Environments)
	duration := time.Since(start)
	logger.Info().
		Int("environment_count", environmentCount).
		Dur("duration", duration).
		Msg("Successfully fetched environments")

	// Use the official API response directly
	responseJSON, err := json.MarshalIndent(chainResponse, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal response")
		return nil, fmt.Errorf("failed to marshal environments response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseJSON),
			},
		},
	}, nil
}
