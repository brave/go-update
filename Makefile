.PHONY: all build test lint clean

all: lint test build

# Even though 'gzhttp' does not use any assembly or unsafe code, it belongs to
# github.com/klauspost/compress, which references these features in its zstd
# implementation.
#
# To stay on the safe side and provide a pure Go implementation, these features
# have been explicitly disabled by setting 'noasm' and 'nounsafe' build tags.
build:
	env AWS_REGION=us-west-2 CGO_ENABLED=0 GOOS=linux go build -tags=noasm,nounsafe -a -o main .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f go-update
