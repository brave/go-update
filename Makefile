.PHONY: all build test lint clean

all: lint test build

# Even though 'gzhttp' does not use any assembly or unsafe code, it belongs to
# github.com/klauspost/compress, which uses these features in its zstd
# implementation.
#
# To ensure a pure Go implementation, assembly and unsafe code have been
# explicitly disabled using the 'noasm' and 'nounsafe' build tags.
build:
	env AWS_REGION=us-west-2 CGO_ENABLED=0 GOOS=linux go build -tags=noasm,nounsafe -a -o main .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f go-update
