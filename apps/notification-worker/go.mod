module github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/notification-worker

go 1.23

require (
	github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require (
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform => ../../pkg/platform
