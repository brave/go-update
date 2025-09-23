FROM golang:1.25 AS builder
WORKDIR /go/src/app

COPY . .
RUN /usr/bin/make build

FROM alpine:latest AS app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/app .
CMD ["./main"]
