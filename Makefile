colon := :


build:
	docker run --rm -v $$PWD$(colon)/app -w /app -e GOPATH=/app/gopath golang$(colon)1.10 go get -d ./...


run:
	docker run --rm -v $$PWD$(colon)/app -w /app -e GOPATH=/app/gopath golang$(colon)1.10 go run main.go postprocessor.go parser.go

run-local:
	go run main.go postprocessor.go parser.go

test:
	docker-compose run app go test -v ./...


.PHONY: run test build