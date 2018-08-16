# Brave component update server written in Go

`go-update` implements a [component update server](https://developer.chrome.com/apps/autoupdate) for use in brave-core written in Go.

The intended audience for this server is all users of brave-core.

The server is only intended to support a small number of extensions that Brave handles ourselves.

The component update server supports 2 types of requests both at the same endpoint `/extensions`

1) The `POST /extensions` endpoint uses an XML schema for the request and the response.  Samples can be found in the tests.
2) The `GET /extensions` endpoint uses URL query parameters and responds with a similar XML schema. Samples can also be found in the tests.

This server is compatible with Google's component update server, so it is a drop-in replacement to handle the requests coming from Chromium.

When there is only a single extension requested, and if we do not support the extension ourselves, we will redirect the request to Google's component updater to handle the request.

This server also serves as a filter so Brave can blacklist any extension before it has a chance to redirect to Google's component updater.


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
