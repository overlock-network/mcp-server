# Overlock MCP Server

Model Context Protocol server for Overlock Network providers.

## Usage

### Local

```bash
# Build and run
make build
make run

# Testing
make test           # Run all tests
make test-unit      # Run unit tests only
make test-e2e       # Run E2E tests only

# Clean
make clean
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Testing with Docker
make docker-test  
```

