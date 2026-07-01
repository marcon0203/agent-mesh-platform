module github.com/agentmesh/billing-service

go 1.25.0

require github.com/agentmesh/shared v0.0.0

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/rabbitmq/amqp091-go v1.12.0 // indirect
)

replace github.com/agentmesh/shared => ../shared
