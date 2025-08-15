package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"overlock-mcp-server/internal/schema"
	"overlock-mcp-server/pkg/config"
	"overlock-mcp-server/pkg/handler"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	overlockv1beta1 "github.com/overlock-network/api/go/node/overlock/crossplane/v1beta1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// validateConnection performs a simple health check on the gRPC connection
func validateConnection(ctx context.Context, queryClient overlockv1beta1.QueryClient) error {
	// Try to make a simple query to validate the connection
	req := &overlockv1beta1.QueryListProviderRequest{
		Pagination: &query.PageRequest{
			Limit:  1,
			Offset: 0,
		},
	}

	// Use a shorter timeout for validation
	valCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := queryClient.ListProvider(valCtx, req)
	return err
}

func startHTTPServer(cfg *config.Config, queryClient overlockv1beta1.QueryClient, grpcConn *grpc.ClientConn) error {
	impl := &mcp.Implementation{
		Name:    "overlock-providers-server",
		Version: "1.0.0",
	}

	srv := mcp.NewServer(impl, nil)

	// Register get-providers tool
	providersTool := &mcp.Tool{
		Name:        "get-providers",
		Description: "Get list of all registered providers in the Overlock Network with optional filtering and pagination",
		InputSchema: schema.CreateProvidersToolInputSchema(),
	}
	providersHandler := handler.NewProvidersHandler(queryClient, cfg.APITimeout)
	mcp.AddTool(srv, providersTool, providersHandler.Handle)

	environmentTool := &mcp.Tool{
		Name:        "show-environment",
		Description: "Get detailed information for a specific environment by its ID",
		InputSchema: schema.CreateEnvironmentToolInputSchema(),
	}
	environmentHandler := handler.NewEnvironmentHandler(queryClient, cfg.APITimeout)
	mcp.AddTool(srv, environmentTool, environmentHandler.Handle)

	// Create the HTTP handler for MCP
	httpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{})

	// Set up HTTP server
	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpHandler,
	}

	if grpcConn != nil {
		log.Info().Str("grpc_url", cfg.OverlockGRPCURL).Msg("Connected to Overlock blockchain")
		log.Info().Msg("Ready to serve Overlock Network data")
	} else {
		log.Warn().Msg("Starting server without gRPC connection - some functionality may be limited")
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", cfg.HTTPAddr).Msg("Starting MCP HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down HTTP server...")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server first
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown failed")
	} else {
		log.Info().Msg("HTTP server stopped")
	}

	// Close gRPC connection
	if grpcConn != nil {
		if err := grpcConn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close gRPC connection")
		} else {
			log.Info().Msg("gRPC connection closed")
		}
	}
	return nil
}

func main() {
	// Initialize structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level based on debug flag
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcConn, err := grpc.NewClient(
		cfg.OverlockGRPCURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Warn().Err(err).Str("grpc_url", cfg.OverlockGRPCURL).Msg("Failed to connect to gRPC server")
		grpcConn = nil
	}

	var queryClient overlockv1beta1.QueryClient
	if grpcConn != nil {
		// Validate connection
		queryClient = overlockv1beta1.NewQueryClient(grpcConn)
		if err := validateConnection(ctx, queryClient); err != nil {
			log.Warn().Err(err).Msg("Failed to validate gRPC connection")
			queryClient = nil
		}
	}

	// Start HTTP server
	if err := startHTTPServer(cfg, queryClient, grpcConn); err != nil {
		log.Fatal().Err(err).Msg("HTTP server error")
	}
}
