# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)


APP             := takiacademy
BASE_IMAGE_NAME := adembc/video-progress
SERVICE_NAME    := video-progress
VERSION         := 0.0.1
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)


# Building containers

all: service

service:
	docker build \
		-f docker/Dockerfile \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

liveness:
	curl -il http://localhost:3001/v1/api/liveness


readiness:
	curl -il http://localhost:3001/v1/api/readiness


# ==============================================================================
# Running tests within the local computer

test-race:
	CGO_ENABLED=1 go test -race -count=1 ./...

test:
	CGO_ENABLED=0 go test -count=1 ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

test-all: test lint vuln-check

test-all-race: test-race lint vuln-check

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all
# ==============================================================================

# Run the application
run:
	go run cmd/main.go

run-help:
	go run cmd/main.go --help
#lint:
#	golangci-lint run

watch:
	@nodemon --exec go run cmd/main.go --signal SIGTERM

build:
	GOOS=linux GOARCH=amd64 go build -o main cmd/main.go
build-mac:
	go build cmd/main.go
binary:
	./main --cron-schedule "*/4 *  * * *"