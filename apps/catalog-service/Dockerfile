FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY pkg/platform ./pkg/platform
COPY apps/catalog-service ./apps/catalog-service
WORKDIR /src/apps/catalog-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog-service ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/catalog-service ./catalog-service
USER app
ENTRYPOINT ["./catalog-service"]
