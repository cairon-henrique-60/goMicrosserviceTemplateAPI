module github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/gateway

go 1.23

require (
	github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.1.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform => ../../pkg/platform
