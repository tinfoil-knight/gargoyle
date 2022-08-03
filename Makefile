build:
	@echo "> Building binary"
	go build -o bin/gargoyle .

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
