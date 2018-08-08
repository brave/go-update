FROM golang:1.10

EXPOSE 8192

ENV GOPATH /go
WORKDIR /go/src/github.com/brave/go-update/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY ./ ./
RUN dep ensure
RUN go install
ENTRYPOINT [ "/go/bin/go-update" ]
