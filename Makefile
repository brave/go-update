.PHONY: all build test lint clean

all: lint test build

build:
	env AWS_REGION=us-west-2 CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

test:
	go test -v ./...

lint:
	golangci-lint run -E gofmt -E golint --exclude-use-default=false

clean:
	rm -f go-update
