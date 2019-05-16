FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git gcc libc-dev
WORKDIR /go-payments-api
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./
ENV GO111MODULE=on
WORKDIR /go-payments-api/cmd
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/go-payments-api

FROM golang:alpine
COPY --from=builder /go/bin/go-payments-api /usr/local/bin/go-payments-api
RUN mkdir -p /etc/go-payments-api/schema
COPY --from=builder /go-payments-api/schema/* /etc/go-payments-api/schema/
CMD ["/usr/local/bin/go-payments-api", "--metrics=true", "--repo-migrations=/etc/go-payments-api/schema"]
