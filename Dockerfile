FROM golang:1.25 AS builder
WORKDIR /go/src/app

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod /usr/bin/make build

FROM alpine:latest AS app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/app .
CMD ["./main"]
