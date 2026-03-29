module tmf/services/demo-ui/bff

go 1.26.1

require (
	github.com/auth0/go-jwt-middleware/v2 v2.3.1
	github.com/gorilla/websocket v1.5.3
	github.com/rabbitmq/amqp091-go v1.10.0
	tmf/pkg v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
)

replace tmf/pkg => ../../../pkg
