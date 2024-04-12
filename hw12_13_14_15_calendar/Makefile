BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
BIN_MIGRATION := "./bin/migration"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build-calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-sched:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

build-sender:
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

generate:
	go generate ./...

run-calend: build-calendar
	$(BIN) -config ./configs/calendar_config.yaml

run-sched: build-sched
	$(BIN_SCHEDULER) -config ./configs/scheduler_config.yaml

run-sender: build-sender
	$(BIN_SENDER) -config ./configs/sender_config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: install-lint-deps
	golangci-lint run ./...

build-migrations:
	go build -v -o $(BIN_MIGRATION) ./cmd/migration

run-migrations: build-migrations
	$(BIN_MIGRATION) -config ./configs/calendar_config.yaml

.PHONY: build build-calendar build-sched build-sender generate run-calendar run-sched run-sender build-img run-img version test lint build-migrations run-migrations
