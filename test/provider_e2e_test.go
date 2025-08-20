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

var _ = Describe("Provider E2E Test", func() {
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

	Describe("Show provider tool", func() {
		Context("when called with valid ID", func() {
			It("should return provider details", func() {
				params := &mcp.CallToolParams{
					Name: "show-provider",
					Arguments: map[string]interface{}{
						"id": 1,
					},
				}

				result, err := providersHandler.HandleShow(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryShowProviderResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Provider).ToNot(BeNil())
				Expect(response.Provider.Creator).To(Equal("overlock1abc123def456ghi789jkl012mno345pqr678stu901vwx"))
				Expect(response.Provider.Id).To(Equal(uint64(1)))
				Expect(response.Provider.Ip).To(Equal("192.168.1.100"))
				Expect(response.Provider.Port).To(Equal(uint32(8080)))
				Expect(response.Provider.CountryCode).To(Equal("US"))
				Expect(response.Provider.EnvironmentType).To(Equal("production"))
				Expect(response.Provider.Availability).To(Equal("available"))
			})
		})

		Context("when called with non-existent ID", func() {
			It("should return provider not found message", func() {
				params := &mcp.CallToolParams{
					Name: "show-provider",
					Arguments: map[string]interface{}{
						"id": 999,
					},
				}

				result, err := providersHandler.HandleShow(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())
				Expect(textContent.Text).To(ContainSubstring("Provider with ID '999' not found"))
			})
		})

		Context("when called without ID", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name:      "show-provider",
					Arguments: nil,
				}

				result, err := providersHandler.HandleShow(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		Context("when called with invalid ID", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name: "show-provider",
					Arguments: map[string]interface{}{
						"id": 0,
					},
				}

				result, err := providersHandler.HandleShow(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		Context("when called with invalid ID type", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name: "show-provider",
					Arguments: map[string]interface{}{
						"id": "invalid", // Invalid type
					},
				}

				result, err := providersHandler.HandleShow(ctx, session, params)
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
				Name: "show-provider",
				Arguments: map[string]interface{}{
					"id": 1,
				},
			}

			result, err := nilHandler.HandleShow(ctx, session, params)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Content).To(HaveLen(1))

			textContent, ok := result.Content[0].(*mcp.TextContent)
			Expect(ok).To(BeTrue())
			Expect(textContent.Text).To(ContainSubstring("gRPC connection to blockchain is not available"))
		})
	})
})