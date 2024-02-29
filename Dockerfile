FROM golang:1.22@sha256:7b297d9abee021bab9046e492506b3c2da8a3722cbf301653186545ecc1e00bb as builder
WORKDIR /go/src/app

COPY . .
RUN /usr/bin/make build

FROM alpine:latest as app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/app .
CMD ["./main"]
