FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o overlock-mcp-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/overlock-mcp-server .

RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENV MCP_HTTP_ADDR=0.0.0.0:8080
ENV OVERLOCK_GRPC_URL=host.docker.internal:9090
ENV OVERLOCK_API_TIMEOUT=30s

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./overlock-mcp-server"]