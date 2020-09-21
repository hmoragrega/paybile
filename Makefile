export DB_IMAGE ?= postgres:9.6
export DB_USER ?= paybile
export DB_PASS ?= paybilepw
export DB_DSN ?= postgresql://${DB_USER}:${DB_PASS}@127.0.0.1:5432/paybile?sslmode=disable

DOCKER_COMPOSE_FLAGS ?= -f ./deployments/docker/docker-compose.yml
DOCKER_COMPOSE ?= docker-compose -p paybile ${DOCKER_COMPOSE_FLAGS}

.PHONY: up
# Build and start up the development environment as daemon.
up:
	${DOCKER_COMPOSE} up -d

.PHONY: build
# Build the images fo the local environment.
build:
	${DOCKER_COMPOSE} build

.PHONY: stop
# Stops the development environment.
stop:
	${DOCKER_COMPOSE} stop

.PHONY: down
# Destroys the development environment.
down:
	${DOCKER_COMPOSE} down

.PHONY: api
# Builds and run the API.
api:
	go run ./cmd/service -db-dsn ${DB_DSN}

.PHONY: migrate-up
# Migrates up the DB schema.
migrate-up: up
	go run ./cmd/migrator -db-dsn ${DB_DSN} -direction up

.PHONY: migrate-down
# Migrates down the DB schema.
migrate-down: up
	go run ./cmd/migrator -db-dsn ${DB_DSN} -direction down

.PHONY: fixtures
# Migrates up the DB schema and reloads the fixtures.
fixtures: migrate-down
	go run ./cmd/migrator -db-dsn ${DB_DSN} -direction up -fixtures true

.PHONY: unit
# Runs unit tests.
unit:
	go test -v -race ./...

.PHONY: test
# Runs all tests.
test:
	go test -v -race -tags integration ./...
