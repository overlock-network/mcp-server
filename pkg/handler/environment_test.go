package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	overlockv1beta1 "github.com/overlock-network/api/go/node/overlock/crossplane/v1beta1"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewEnvironmentHandler(t *testing.T) {
	mockClient := &MockQueryClient{}
	timeout := 30 * time.Second

	handler := NewEnvironmentHandler(mockClient, timeout)

	assert.NotNil(t, handler)
	assert.Equal(t, mockClient, handler.chainClient)
	assert.Equal(t, timeout, handler.timeout)
	assert.NotNil(t, handler.circuitBreaker)
	assert.Equal(t, "blockchain-client-env", handler.circuitBreaker.Name())
}

func TestEnvironmentHandler_Handle_Success(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewEnvironmentHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryShowEnvironmentResponse{
		Environment: &overlockv1beta1.Environment{
			Id:      1,
			Creator: "test-creator",
		},
	}

	mockClient.On("ShowEnvironment", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var response overlockv1beta1.QueryShowEnvironmentResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Environment)
	assert.Equal(t, uint64(1), response.Environment.Id)
	assert.Equal(t, "test-creator", response.Environment.Creator)

	mockClient.AssertExpectations(t)
}

func TestEnvironmentHandler_Handle_NilClient(t *testing.T) {
	handler := NewEnvironmentHandler(nil, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestEnvironmentHandler_Handle_ValidationError_MissingID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewEnvironmentHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name:      "show-environment",
		Arguments: map[string]interface{}{},
	}

	result, err := handler.Handle(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestEnvironmentHandler_Handle_ValidationError_InvalidID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewEnvironmentHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": "invalid-string-id",
		},
	}

	result, err := handler.Handle(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestEnvironmentHandler_Handle_EnvironmentNotFound(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewEnvironmentHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Response with nil environment indicates not found
	expectedResponse := &overlockv1beta1.QueryShowEnvironmentResponse{
		Environment: nil,
	}

	mockClient.On("ShowEnvironment", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": 999,
		},
	}

	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Environment with ID '999' not found")

	mockClient.AssertExpectations(t)
}

func TestEnvironmentHandler_Handle_CircuitBreakerOpen(t *testing.T) {
	handler := NewEnvironmentHandler(nil, 30*time.Second)

	handler.circuitBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "test-breaker",
		MaxRequests: 0,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return true
		},
	})

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestEnvironmentHandler_Handle_WithValidID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewEnvironmentHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryShowEnvironmentResponse{
		Environment: &overlockv1beta1.Environment{
			Id:      123,
			Creator: "production-creator",
		},
	}

	envID := uint64(123)
	mockClient.On("ShowEnvironment", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req *overlockv1beta1.QueryShowEnvironmentRequest) bool {
		return req.Id == envID
	})).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-environment",
		Arguments: map[string]interface{}{
			"id": 123,
		},
	}

	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var response overlockv1beta1.QueryShowEnvironmentResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Environment)
	assert.Equal(t, envID, response.Environment.Id)

	mockClient.AssertExpectations(t)
}
