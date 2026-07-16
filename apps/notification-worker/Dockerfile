FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY pkg/platform ./pkg/platform
COPY apps/notification-worker ./apps/notification-worker
WORKDIR /src/apps/notification-worker
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notification-worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/notification-worker ./notification-worker
USER app
ENTRYPOINT ["./notification-worker"]
