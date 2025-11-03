.PHONY: test lint fmt clean

test:
	go test ./... -v

lint:
	golangci-lint run

fmt:
	go fmt ./...

clean:
	go clean ./...

