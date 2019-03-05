build:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 go get -d ./... && go build

build-local:
	go get -d ./... && go build

run:
<<<<<<< HEAD
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 go run main.go postprocessor.go parser.go -i files/test.txt

run-local:
	go run main.go postprocessor.go parser.go -i files/test.txt
=======
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 ./gofigure config.txt

run-local:
	./gofigure config.txt
>>>>>>> tests

.PHONY: run run-local build
