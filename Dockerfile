FROM golang:1.10 as dep

ENV DEP_VERSION 0.5.0
ENV DEP_SHA256SUM 287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v$DEP_VERSION/dep-linux-amd64
RUN echo "$DEP_SHA256SUM  /usr/local/bin/dep" | shasum -a 256 -c
RUN chmod +x /usr/local/bin/dep
RUN go get -u github.com/golangci/golangci-lint/cmd/golangci-lint


FROM golang:1.10 as builder
WORKDIR /go/src/app

COPY --from=dep /usr/local/bin/dep /usr/local/bin/dep
COPY --from=dep /go/bin/golangci-lint /bin/golangci-lint
COPY . .
RUN dep ensure
RUN /usr/bin/make build

FROM alpine:latest as app
RUN apk add --update ca-certificates # Certificates for SSL
COPY --from=builder  /go/src/app .
CMD ["./main"]
