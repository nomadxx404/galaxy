FROM golang:1.25.6-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git sed

ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o companies-service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/companies-service .

RUN mkdir -p ./static
COPY openapi.yaml ./static/swagger.json

EXPOSE 8080
CMD ["./companies-service"]
