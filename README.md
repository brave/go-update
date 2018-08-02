# Brave component update server written in Go

## Dependencies

- Install Go 1.10 or later.
- `dep` is used to install the Go dependencies.
- `go get -u github.com/golangci/golangci-lint/cmd/golangci-lint`

## Setup

```
go get -d github.com/brave/go-update
cd ~/go/src/github.com/brave/go-update
dep ensure
```

## Run lint:

`make lint`

## Run tests:

`make test`

## Run go-update:

`make`
