FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git gcc libc-dev
ENV GOBIN /go/bin
WORKDIR $GOPATH/src/github.com/mfamador/go-payments-api
ADD . .
WORKDIR $GOPATH/src/github.com/mfamador/go-payments-api/cmd
RUN go get
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/go-payments-api

FROM golang:alpine
COPY --from=builder /go/bin/go-payments-api /usr/local/bin/go-payments-api
RUN mkdir -p /etc/go-payments-api/schema
COPY --from=builder $GOPATH/src/github.com/mfamador/go-payments-api/schema/* /etc/go-payments-api/schema/
CMD ["/usr/local/bin/go-payments-api", "--metrics=true", "--repo-migrations=/etc/go-payments-api/schema"]
