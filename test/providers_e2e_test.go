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

var _ = Describe("Providers E2E Test", func() {
	var (
		mockClient       *mocks.MockQueryClient
		providersHandler *handler.ProvidersHandler
		ctx              context.Context
		session          *mcp.ServerSession
	)

	BeforeEach(func() {
		ctx = context.Background()
		session = &mcp.ServerSession{}

		// Setup mock client with test data
		testDataDir, err := filepath.Abs("testdata")
		Expect(err).ToNot(HaveOccurred())

		mockClient = mocks.NewMockQueryClient(testDataDir)
		providersHandler = handler.NewProvidersHandler(mockClient, 30*time.Second)
	})

	Describe("Get providers tool", func() {
		Context("when called without arguments", func() {
			It("should return all providers", func() {
				params := &mcp.CallToolParams{
					Name:      "get-providers",
					Arguments: nil,
				}

				result, err := providersHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListProviderResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Providers).To(HaveLen(2))
				Expect(response.Providers[0].Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Providers[1].Creator).To(Equal("overlock2xyz789abc012def345ghi678jkl901mno234pqr567stu"))
			})
		})

		Context("when called with creator filter", func() {
			It("should return filtered providers", func() {
				params := &mcp.CallToolParams{
					Name: "get-providers",
					Arguments: map[string]interface{}{
						"creator": "overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx",
					},
				}

				result, err := providersHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListProviderResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Providers).To(HaveLen(1))
				Expect(response.Providers[0].Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Providers[0].Metadata.Name).To(Equal("test-provider-1"))
			})
		})

		Context("when called with pagination", func() {
			It("should return paginated results", func() {
				params := &mcp.CallToolParams{
					Name: "get-providers",
					Arguments: map[string]interface{}{
						"limit":  1,
						"offset": 0,
					},
				}

				result, err := providersHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListProviderResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Providers).To(HaveLen(1))
				Expect(response.Providers[0].Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Pagination.Total).To(Equal(uint64(1)))
			})
		})

		Context("when called with invalid arguments", func() {
			It("should handle empty creator gracefully", func() {
				params := &mcp.CallToolParams{
					Name: "get-providers",
					Arguments: map[string]interface{}{
						"creator": "", // Empty string should be valid (no filtering)
					},
				}

				result, err := providersHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryListProviderResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				// Empty creator means no filtering, should return all providers
				Expect(response.Providers).To(HaveLen(2))
			})

			It("should handle limit over maximum", func() {
				params := &mcp.CallToolParams{
					Name: "get-providers",
					Arguments: map[string]interface{}{
						"limit": 2000, // Over max of 1000
					},
				}

				result, err := providersHandler.Handle(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})
	})

	Describe("Handler with nil client", func() {
		It("should return error message when gRPC client is nil", func() {
			// Create handler with nil client
			nilHandler := handler.NewProvidersHandler(nil, 30*time.Second)

			params := &mcp.CallToolParams{
				Name:      "get-providers",
				Arguments: nil,
			}

			result, err := nilHandler.Handle(ctx, session, params)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Content).To(HaveLen(1))

			textContent, ok := result.Content[0].(*mcp.TextContent)
			Expect(ok).To(BeTrue())
			Expect(textContent.Text).To(ContainSubstring("gRPC connection to blockchain is not available"))
		})
	})
})
