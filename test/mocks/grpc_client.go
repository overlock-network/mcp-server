package mocks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/types/query"
	overlockv1beta1 "github.com/overlock-network/api/go/node/overlock/crossplane/v1beta1"
	"google.golang.org/grpc"
)

// MockQueryClient implements overlockv1beta1.QueryClient for testing
type MockQueryClient struct {
	testDataPath string
}

// NewMockQueryClient creates a new mock query client
func NewMockQueryClient(testDataPath string) *MockQueryClient {
	return &MockQueryClient{
		testDataPath: testDataPath,
	}
}

// ListProvider implements the ListProvider method by returning test data
func (m *MockQueryClient) ListProvider(ctx context.Context, req *overlockv1beta1.QueryListProviderRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryListProviderResponse, error) {
	// Load test data from JSON file
	jsonFile := filepath.Join(m.testDataPath, "providers_response.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var response overlockv1beta1.QueryListProviderResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Apply filtering if creator is specified
	if req.Creator != nil && req.Creator.Value != "" {
		var filteredProviders []overlockv1beta1.Provider
		for _, provider := range response.Providers {
			if provider.Creator == req.Creator.Value {
				filteredProviders = append(filteredProviders, provider)
			}
		}
		response.Providers = filteredProviders
	}

	// Apply pagination
	if req.Pagination != nil {
		offset := int(req.Pagination.Offset)
		limit := int(req.Pagination.Limit)

		if offset >= len(response.Providers) {
			response.Providers = []overlockv1beta1.Provider{}
		} else {
			end := offset + limit
			if end > len(response.Providers) {
				end = len(response.Providers)
			}
			response.Providers = response.Providers[offset:end]
		}

		// Update pagination info
		response.Pagination = &query.PageResponse{
			NextKey: []byte{},
			Total:   uint64(len(response.Providers)),
		}
	}

	return &response, nil
}

// ShowProvider implements the ShowProvider method by returning test data
func (m *MockQueryClient) ShowProvider(ctx context.Context, req *overlockv1beta1.QueryShowProviderRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryShowProviderResponse, error) {
	// Load test data from JSON file
	jsonFile := filepath.Join(m.testDataPath, "provider_response.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var response overlockv1beta1.QueryShowProviderResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Check if the requested ID matches the test data
	if req.Id != 1 {
		// Return empty response for non-matching IDs (provider not found)
		return &overlockv1beta1.QueryShowProviderResponse{
			Provider: nil,
		}, nil
	}

	return &response, nil
}

// ShowEnvironment implements the ShowEnvironment method by returning test data
func (m *MockQueryClient) ShowEnvironment(ctx context.Context, req *overlockv1beta1.QueryShowEnvironmentRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryShowEnvironmentResponse, error) {
	// Load test data from JSON file
	jsonFile := filepath.Join(m.testDataPath, "environment_response.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var response overlockv1beta1.QueryShowEnvironmentResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Check if the requested ID matches the test data
	if req.Id != 1001 {
		// Return empty response for non-matching IDs (environment not found)
		return &overlockv1beta1.QueryShowEnvironmentResponse{
			Environment: nil,
		}, nil
	}

	return &response, nil
}

// ListEnvironment implements the ListEnvironment method by returning test data
func (m *MockQueryClient) ListEnvironment(ctx context.Context, req *overlockv1beta1.QueryListEnvironmentRequest, opts ...grpc.CallOption) (*overlockv1beta1.QueryListEnvironmentResponse, error) {
	// Load test data from JSON file
	jsonFile := filepath.Join(m.testDataPath, "environments_response.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var response overlockv1beta1.QueryListEnvironmentResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Apply filtering if creator is specified
	if req.Creator != "" {
		var filteredEnvironments []overlockv1beta1.Environment
		for _, environment := range response.Environments {
			if environment.Creator == req.Creator {
				filteredEnvironments = append(filteredEnvironments, environment)
			}
		}
		response.Environments = filteredEnvironments
	}

	// Apply pagination
	if req.Pagination != nil {
		offset := int(req.Pagination.Offset)
		limit := int(req.Pagination.Limit)

		if offset >= len(response.Environments) {
			response.Environments = []overlockv1beta1.Environment{}
		} else {
			end := offset + limit
			if end > len(response.Environments) {
				end = len(response.Environments)
			}
			response.Environments = response.Environments[offset:end]
		}

		// Update pagination info
		response.Pagination = &query.PageResponse{
			NextKey: []byte{},
			Total:   uint64(len(response.Environments)),
		}
	}

	return &response, nil
}
