FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY pkg/platform ./pkg/platform
COPY apps/auth-service ./apps/auth-service
WORKDIR /src/apps/auth-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/auth-service ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/auth-service ./auth-service
USER app
ENTRYPOINT ["./auth-service"]
