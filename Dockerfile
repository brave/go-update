FROM golang:1.22 as builder
WORKDIR /go/src/app

COPY . .
RUN /usr/bin/make build

FROM alpine:latest as app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/app .
CMD ["./main"]
