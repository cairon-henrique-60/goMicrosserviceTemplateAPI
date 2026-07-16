FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY pkg/platform ./pkg/platform
COPY apps/gateway ./apps/gateway
WORKDIR /src/apps/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/gateway ./gateway
USER app
ENTRYPOINT ["./gateway"]
