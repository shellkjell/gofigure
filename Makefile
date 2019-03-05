build:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 go get -d ./...


run:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 go run main.go postprocessor.go parser.go

run-local:
	go run main.go postprocessor.go parser.go

.PHONY: run run-local build
