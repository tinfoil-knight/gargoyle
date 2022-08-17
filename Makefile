GOROOT = $(shell go env GOROOT)

build:
	@echo "> Building binary"
	go build -o bin/gargoyle .

build-docker:
	@echo "> Building Docker image"
	docker build -t gargoyle .

gencert:
	@echo "> Generating TLS cert"
	mkdir -p .gargoyle
	cd .gargoyle; go run "$(GOROOT)/src/crypto/tls/generate_cert.go" --rsa-bits=2048 --host=localhost

run:
	@echo "> Starting Gargoyle"
	go run gargoyle.go

test:
	@echo "> Running tests"
	go test -v -race ./...

format:
	@echo "> Formatting the source"
	gofmt -d -e

clean:
	@echo "> Cleaning up"
	go clean -testcache
	rm -rf tmp bin

.PHONY: build run format
