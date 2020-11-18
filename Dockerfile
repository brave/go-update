FROM golang:1.15.5 as builder

ENV DEP_VERSION 0.5.4
ENV DEP_SHA256SUM 40a78c13753f482208d3f4bea51244ca60a914341050c588dad1f00b1acc116c

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v$DEP_VERSION/dep-linux-amd64
RUN echo "$DEP_SHA256SUM  /usr/local/bin/dep" | shasum -a 256 -c
RUN chmod +x /usr/local/bin/dep
RUN mkdir -p /go/src/github.com/brave/go-update

WORKDIR /go/src/github.com/brave/go-update
COPY . .
RUN dep ensure
RUN /usr/bin/make build

FROM alpine:latest as app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/github.com/brave/go-update/main .
CMD ["./main"]
