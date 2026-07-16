FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY pkg/platform ./pkg/platform
COPY apps/order-service ./apps/order-service
WORKDIR /src/apps/order-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/order-service ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/order-service ./order-service
USER app
ENTRYPOINT ["./order-service"]
