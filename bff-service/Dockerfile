FROM golang:1.25.7-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o bff-service ./cmd/main.go

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S appuser && adduser -S appuser -G appuser

COPY --from=builder /app/bff-service .
COPY --from=builder /app/openapi.yaml ./openapi.yaml

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8001
ENTRYPOINT ["./bff-service"]
