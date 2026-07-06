FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /api .
COPY configs/config.example.yaml ./configs/config.yaml
COPY migrations ./migrations

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./api", "-config", "configs/config.yaml"]
