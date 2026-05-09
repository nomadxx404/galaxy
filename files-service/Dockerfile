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
    -o files-service ./cmd/main.go

FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache busybox-extras curl

RUN addgroup -S appuser && adduser -S appuser -G appuser

COPY --from=builder /app/files-service .
COPY openapi.yaml .

RUN mkdir -p ./static
COPY openapi.yaml ./static/swagger.json

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
ENTRYPOINT ["./files-service"]
