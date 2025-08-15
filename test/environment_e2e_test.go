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

var _ = Describe("Environment E2E Test", func() {
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

	Describe("Show environment tool", func() {
		Context("when called with valid environment ID", func() {
			It("should return environment details", func() {
				params := &mcp.CallToolParams{
					Name: "show-environment",
					Arguments: map[string]interface{}{
						"id": 1001,
					},
				}

				result, err := environmentHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryShowEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environment).ToNot(BeNil())
				Expect(response.Environment.Id).To(Equal(uint64(1001)))
				Expect(response.Environment.Creator).To(Equal("overlock1test123abc456def789creator"))
				Expect(response.Environment.Provider).To(Equal(uint64(2001)))
				Expect(response.Environment.Metadata.Name).To(Equal("production-environment"))
			})
		})

		Context("when called with non-existent environment ID", func() {
			It("should return environment not found message", func() {
				params := &mcp.CallToolParams{
					Name: "show-environment",
					Arguments: map[string]interface{}{
						"id": 9999,
					},
				}

				result, err := environmentHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Content).To(HaveLen(1))

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())
				Expect(textContent.Text).To(ContainSubstring("Environment with ID '9999' not found"))
			})
		})

		Context("when called without required ID parameter", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name:      "show-environment",
					Arguments: map[string]interface{}{},
				}

				result, err := environmentHandler.Handle(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		Context("when called with invalid ID parameter", func() {
			It("should return validation error", func() {
				params := &mcp.CallToolParams{
					Name: "show-environment",
					Arguments: map[string]interface{}{
						"id": "invalid-string",
					},
				}

				result, err := environmentHandler.Handle(ctx, session, params)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		Context("when called with valid ID but different format", func() {
			It("should handle the ID parameter correctly", func() {
				params := &mcp.CallToolParams{
					Name: "show-environment",
					Arguments: map[string]interface{}{
						"id": 1001, // This matches our test data
					},
				}

				result, err := environmentHandler.Handle(ctx, session, params)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())

				textContent, ok := result.Content[0].(*mcp.TextContent)
				Expect(ok).To(BeTrue())

				var response overlockv1beta1.QueryShowEnvironmentResponse
				err = json.Unmarshal([]byte(textContent.Text), &response)
				Expect(err).ToNot(HaveOccurred())

				Expect(response.Environment).ToNot(BeNil())
				Expect(response.Environment.Id).To(Equal(uint64(1001)))
				Expect(response.Environment.Metadata.Annotations).To(ContainSubstring("region"))
				Expect(response.Environment.Metadata.Annotations).To(ContainSubstring("us-east-1"))
			})
		})
	})

	Describe("Handler with nil client", func() {
		It("should return error message when gRPC client is nil", func() {
			// Create handler with nil client
			nilHandler := handler.NewEnvironmentHandler(nil, 30*time.Second)

			params := &mcp.CallToolParams{
				Name: "show-environment",
				Arguments: map[string]interface{}{
					"id": 1001,
				},
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
