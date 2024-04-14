BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
BIN_MIGRATION := "./bin/migration"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.PHONY: build-calendar
build-calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

.PHONY: build-sched
build-sched:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

.PHONY: build-sender
build-sender:
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

.PHONY: build
build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

.PHONY: generate
generate:
	go generate ./...

.PHONY: run-calend
run-calend: build-calendar
	$(BIN) -config ./configs/calendar_config.yaml

.PHONY: run-sched
run-sched: build-sched
	$(BIN_SCHEDULER) -config ./configs/scheduler_config.yaml

.PHONY: run-sender
run-sender: build-sender
	$(BIN_SENDER) -config ./configs/sender_config.yaml

.PHONY: build-img
build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

.PHONY: run-img
run-img: build-img
	docker run $(DOCKER_IMG)

.PHONY: version
version: build
	$(BIN) version

.PHONY: test
test:
	go test -race ./internal/... ./pkg/...

.PHONY: install-lint-deps
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

.PHONY: lint
lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build-migrations
build-migrations:
	go build -v -o $(BIN_MIGRATION) ./cmd/migration

.PHONY: run-migrations
run-migrations: build-migrations
	$(BIN_MIGRATION) -config ./configs/calendar_config.yaml

.PHONY: up
up:
	docker-compose -f deployments/docker-compose.yaml up -d

.PHONY: down
down:
	docker-compose -f deployments/docker-compose.yaml down

.PHONY: integration-tests
integration-tests:
	docker-compose -f deployments/docker-compose_test.yaml up --abort-on-container-exit --exit-code-from run-test run-test
	docker-compose -f deployments/docker-compose_test.yaml down