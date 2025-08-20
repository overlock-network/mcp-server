package test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"overlock-mcp-server/pkg/handler"
	"overlock-mcp-server/test/mocks"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	overlockv1beta1 "github.com/overlock-network/api/go/node/overlock/crossplane/v1beta1"
)

var _ = Describe("Environments E2E Test", func() {
	var (
		mockClient         *mocks.MockQueryClient
		environmentHandler *handler.EnvironmentHandler
		ctx                context.Context
		session            *mcp.ServerSession
	)

	BeforeEach(func() {
		ctx = context.Background()
		session = &mcp.ServerSession{}

		// Setup mock client with test data
		testDataDir, err := filepath.Abs("testdata")
		Expect(err).ToNot(HaveOccurred())

		mockClient = mocks.NewMockQueryClient(testDataDir)
		environmentHandler = handler.NewEnvironmentHandler(mockClient, 30*time.Second)
	})

	Describe("Get environments tool", func() {
		Context("when called without arguments", func() {
			It("should return all environments", func() {
				params := &mcp.CallToolParams{
					Name:      "get-environments",
					Arguments: nil,
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(2))
				Expect(response.Environments[0].Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Environments[1].Creator).To(Equal("overlock1xyz789ghi456def123abc012mno345pqr678stu901vwx"))
				Expect(response.Environments[0].Metadata.Name).To(Equal("test-environment-1"))
				Expect(response.Environments[1].Metadata.Name).To(Equal("test-environment-2"))
			})
		})

		Context("when called with creator filter", func() {
			It("should return filtered environments", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"creator": "overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx",
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(1))
				Expect(response.Environments[0].Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Environments[0].Metadata.Name).To(Equal("test-environment-1"))
				Expect(response.Environments[0].Metadata.Annotations).To(ContainSubstring("Test environment for E2E testing"))
			})
		})

		Context("when called with pagination", func() {
			It("should return paginated results", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"limit":  1,
						"offset": 0,
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(1))
				Expect(response.Environments[0].Id).To(Equal(uint64(1001)))
				Expect(response.Environments[0].Provider).To(Equal(uint64(1)))
			})
		})

		Context("when called with empty creator filter", func() {
			It("should return all environments", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"creator": "",
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(2))
			})
		})

		Context("when called with invalid limit", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"limit": 2000, // Exceeds maximum limit of 1000
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		Context("when called with valid environment metadata", func() {
			It("should return environment with proper metadata structure", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"limit": 1,
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(1))
				env := response.Environments[0]
				
				Expect(env.Metadata).ToNot(BeNil())
				Expect(env.Metadata.Name).To(Equal("test-environment-1"))
				Expect(env.Metadata.Annotations).To(ContainSubstring("region"))
				Expect(env.Metadata.Annotations).To(ContainSubstring("us-east-1"))
				Expect(env.Metadata.Annotations).To(ContainSubstring("development"))
				Expect(env.Metadata.Annotations).To(ContainSubstring("Test environment for E2E testing"))
				
				Expect(env.Provider).To(Equal(uint64(1)))
				Expect(env.Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
			})
		})

		Context("when called with non-existent creator filter", func() {
			It("should return empty results", func() {
				params := &mcp.CallToolParams{
					Name: "get-environments",
					Arguments: map[string]interface{}{
						"creator": "overlock1nonexistent123456789",
					},
				}

				result, err := environmentHandler.HandleList(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environments).To(HaveLen(0))
			})
		})
	})

	Describe("Handler with nil client", func() {
		It("should return error message when gRPC client is nil", func() {
			// Create handler with nil client
			nilHandler := handler.NewEnvironmentHandler(nil, 30*time.Second)

			params := &mcp.CallToolParams{
				Name:      "get-environments",
				Arguments: map[string]interface{}{},
			}

			result, err := nilHandler.HandleList(ctx, session, params)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Content).To(HaveLen(1))

			textContent, ok := result.Content[0].(*mcp.TextContent)
			Expect(ok).To(BeTrue())
			Expect(textContent.Text).To(ContainSubstring("gRPC connection to blockchain is not available"))
		})
	})
})