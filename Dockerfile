FROM golang:1.15 as builder
RUN mkdir -p /go/src/github.com/brave/go-update
WORKDIR /go/src/github.com/brave/go-update

COPY . .
RUN /usr/bin/make build

FROM alpine:latest as app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/github.com/brave/go-update/main .
CMD ["./main"]
