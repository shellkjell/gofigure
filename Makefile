build:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 go get -d ./... && go build

build-local:
	go get -d ./... && go build

run:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 ./gofigure -i config.fig

run-local:
	./gofigure -i config.fig

.PHONY: run run-local build
