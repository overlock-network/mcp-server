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
	"google.golang.org/grpc"
)

type MockQueryClient struct {
	mock.Mock
}

func (m *MockQueryClient) ListProvider(ctx context.Context, req *overlockv1beta1.QueryListProviderRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryListProviderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*overlockv1beta1.QueryListProviderResponse), args.Error(1)
}

func (m *MockQueryClient) ShowProvider(ctx context.Context, req *overlockv1beta1.QueryShowProviderRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryShowProviderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*overlockv1beta1.QueryShowProviderResponse), args.Error(1)
}

func (m *MockQueryClient) ShowEnvironment(ctx context.Context, req *overlockv1beta1.QueryShowEnvironmentRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryShowEnvironmentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*overlockv1beta1.QueryShowEnvironmentResponse), args.Error(1)
}

func (m *MockQueryClient) ListEnvironment(ctx context.Context, req *overlockv1beta1.QueryListEnvironmentRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryListEnvironmentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*overlockv1beta1.QueryListEnvironmentResponse), args.Error(1)
}

func TestNewProvidersHandler(t *testing.T) {
	mockClient := &MockQueryClient{}
	timeout := 30 * time.Second

	handler := NewProvidersHandler(mockClient, timeout)

	assert.NotNil(t, handler)
	assert.Equal(t, mockClient, handler.chainClient)
	assert.Equal(t, timeout, handler.timeout)
	assert.NotNil(t, handler.circuitBreaker)
	assert.Equal(t, "blockchain-client-providers", handler.circuitBreaker.Name())
}

// Tests for HandleList (get-providers functionality)

func TestProvidersHandler_HandleList_Success(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryListProviderResponse{
		Providers: []overlockv1beta1.Provider{
			{
				Id:      1,
				Creator: "test-creator",
			},
		},
	}

	mockClient.On("ListProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "get-providers",
		Arguments: map[string]interface{}{
			"limit": 10,
		},
	}

	result, err := handler.HandleList(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var response overlockv1beta1.QueryListProviderResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.Len(t, response.Providers, 1)

	mockClient.AssertExpectations(t)
}

func TestProvidersHandler_HandleList_NilClient(t *testing.T) {
	handler := NewProvidersHandler(nil, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name:      "get-providers",
		Arguments: nil,
	}

	result, err := handler.HandleList(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestProvidersHandler_HandleList_ValidationError(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "get-providers",
		Arguments: map[string]interface{}{
			"limit": 2000, // Over max of 1000
		},
	}

	result, err := handler.HandleList(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestProvidersHandler_HandleList_DefaultValues(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryListProviderResponse{
		Providers: []overlockv1beta1.Provider{},
	}

	mockClient.On("ListProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name:      "get-providers",
		Arguments: nil, // Should apply default values
	}

	result, err := handler.HandleList(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)

	mockClient.AssertExpectations(t)
}

func TestProvidersHandler_HandleList_CircuitBreakerOpen(t *testing.T) {
	handler := NewProvidersHandler(nil, 30*time.Second)

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
		Name:      "get-providers",
		Arguments: nil,
	}

	result, err := handler.HandleList(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestProvidersHandler_HandleList_WithCreatorFilter(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryListProviderResponse{
		Providers: []overlockv1beta1.Provider{},
	}

	mockClient.On("ListProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "get-providers",
		Arguments: map[string]interface{}{
			"creator": "test-creator-address",
		},
	}

	result, err := handler.HandleList(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)

	mockClient.AssertExpectations(t)
}

// Tests for HandleShow (show-provider functionality)

func TestProvidersHandler_HandleShow_Success(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryShowProviderResponse{
		Provider: &overlockv1beta1.Provider{
			Id:      1,
			Creator: "overlock1test123",
		},
	}

	mockClient.On("ShowProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var response overlockv1beta1.QueryShowProviderResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Provider)
	assert.Equal(t, uint64(1), response.Provider.Id)
	assert.Equal(t, "overlock1test123", response.Provider.Creator)

	mockClient.AssertExpectations(t)
}

func TestProvidersHandler_HandleShow_NilClient(t *testing.T) {
	handler := NewProvidersHandler(nil, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestProvidersHandler_HandleShow_ValidationError_MissingID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name:      "show-provider",
		Arguments: map[string]interface{}{},
	}

	result, err := handler.HandleShow(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestProvidersHandler_HandleShow_ValidationError_InvalidID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 0,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestProvidersHandler_HandleShow_ValidationError_InvalidType(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": "invalid", // Invalid type
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestProvidersHandler_HandleShow_ProviderNotFound(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Response with nil provider indicates not found
	expectedResponse := &overlockv1beta1.QueryShowProviderResponse{
		Provider: nil,
	}

	mockClient.On("ShowProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 999,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Provider with ID '999' not found")

	mockClient.AssertExpectations(t)
}

func TestProvidersHandler_HandleShow_CircuitBreakerOpen(t *testing.T) {
	handler := NewProvidersHandler(nil, 30*time.Second)

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
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 1,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "gRPC connection to blockchain is not available")
}

func TestProvidersHandler_HandleShow_WithValidID(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryShowProviderResponse{
		Provider: &overlockv1beta1.Provider{
			Id:      123,
			Creator: "overlock1production123",
		},
	}

	testID := uint64(123)
	mockClient.On("ShowProvider", mock.AnythingOfType("*context.timerCtx"), mock.MatchedBy(func(req *overlockv1beta1.QueryShowProviderRequest) bool {
		return req.Id == testID
	})).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name: "show-provider",
		Arguments: map[string]interface{}{
			"id": 123,
		},
	}

	result, err := handler.HandleShow(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var response overlockv1beta1.QueryShowProviderResponse
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Provider)
	assert.Equal(t, testID, response.Provider.Id)
	assert.Equal(t, "overlock1production123", response.Provider.Creator)

	mockClient.AssertExpectations(t)
}

// Test the backward compatibility Handle method

func TestProvidersHandler_Handle_BackwardCompatibility(t *testing.T) {
	mockClient := &MockQueryClient{}
	handler := NewProvidersHandler(mockClient, 30*time.Second)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	expectedResponse := &overlockv1beta1.QueryListProviderResponse{
		Providers: []overlockv1beta1.Provider{},
	}

	mockClient.On("ListProvider", mock.AnythingOfType("*context.timerCtx"), mock.Anything).Return(expectedResponse, nil)

	params := &mcp.CallToolParams{
		Name:      "get-providers",
		Arguments: nil,
	}

	// Test that Handle method still works (routes to HandleList)
	result, err := handler.Handle(ctx, session, params)

	require.NoError(t, err)
	require.NotNil(t, result)

	mockClient.AssertExpectations(t)
}