# Include .env file
include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run the cmd application
.PHONY: run
run:
	@echo 'Start the app in development mode...'
	@go run ./cmd/app -env=dev

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	GOWORK=off go mod vendor

## cover: test coverage
.PHONY: cover
cover:
	@echo 'Test coverage...'
	go test -covermode=count -coverprofile=/tmp/profile.out ./... 

## analyze: analyze the test coverage in your browser
.PHONY: analyze
analyze: cover
	@echo 'Analyue test coverage...'
	go tool cover -html=/tmp/profile.out

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/linux_amd64: build the service
.PHONY: build/linux_amd64
build/linux_amd64: audit
	@echo 'Building cmd/${SERVICE}...'
	go build -ldflags='-s' -o=./bin/${SERVICE} ./cmd/app
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --tags extended -ldflags='-s' -o=./bin/linux_amd64/${SERVICE} ./cmd/app