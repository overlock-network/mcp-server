# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server for the Overlock Network blockchain, written in Go 1.24. The server provides an HTTP API that exposes Overlock Network provider data through the MCP specification, allowing AI assistants to query blockchain provider information.

## Key Commands

### Build and Run
```bash
make build          # Build the binary to bin/overlock-mcp-server
make run            # Build and run the server locally
make clean          # Clean build artifacts
```

### Testing
```bash
make test           # Run all tests (unit + E2E)
make test-unit      # Run unit tests only (pkg/... internal/...)
make test-e2e       # Run E2E tests using Ginkgo (test/...)
```

### Docker Operations
```bash
make docker-build   # Build Docker image
make docker-run     # Build and run in Docker (port 8080)
make docker-test    # Run all tests in Docker containers
```

## Architecture

### Core Components

- **cmd/server/main.go**: Entry point that sets up gRPC client, MCP server, and HTTP endpoints
- **pkg/config/**: Environment-based configuration management with validation
- **pkg/handler/**: Business logic for the `get-providers` MCP tool with circuit breaker protection
- **internal/schema/**: JSON schema definitions for MCP tool input validation

### Key Dependencies

- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` for Model Context Protocol implementation
- **Overlock API**: `github.com/overlock-network/api` for blockchain interaction via gRPC
- **Cosmos SDK**: Used for blockchain query pagination and types
- **Circuit Breaker**: `github.com/sony/gobreaker` for resilience against blockchain service failures
- **Validation**: `github.com/Oudwins/zog` for input parameter validation
- **Logging**: `github.com/rs/zerolog` for structured logging

### Data Flow

1. HTTP requests received by MCP server at `/` endpoint
2. `get-providers` tool calls validated through JSON schema
3. Parameters passed to ProvidersHandler with circuit breaker protection
4. gRPC calls made to Overlock blockchain QueryClient
5. Blockchain response marshaled to JSON and returned via MCP

### Configuration

Environment variables (with defaults):
- `OVERLOCK_GRPC_URL`: Blockchain gRPC endpoint (default: localhost:9090)
- `MCP_HTTP_ADDR`: HTTP server address (default: 127.0.0.1:8080)
- `OVERLOCK_API_TIMEOUT`: API request timeout (default: 30s)
- `DEBUG`: Enable debug logging (default: false)

### Error Handling

- Circuit breaker pattern protects against blockchain service failures
- Graceful degradation when gRPC connection is unavailable
- User-friendly error messages returned instead of technical details
- Structured logging with request context for debugging

### Testing Strategy

- **Unit tests**: Focus on pkg/ and internal/ packages with mocked dependencies
- **E2E tests**: Use Ginkgo/Gomega framework to test full integration flows
- **Test data**: Located in test/testdata/ for consistent test scenarios