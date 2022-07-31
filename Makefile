build:
	@echo "> Building binary"
	go build -o bin/gargoyle .

run:
	@echo "> Starting Gargoyle"
	go run gargoyle.go

format:
	@echo "> Formatting the source"
	gofmt -d -e

clean:
	@echo "> Cleaning up"
	rm -rf tmp bin

.PHONY: build run format
