FROM golang:1.17-alpine as builder
WORKDIR /go/src/github.com/koordinator-sh/koordetector
RUN apk add --update make git bash rsync gcc musl-dev

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# Copy the go source
COPY apis/ apis/
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN GOOS=linux GOARCH=amd64 go build -a -o interference-manager cmd/interference-manager/main.go

FROM alpine:3.12
RUN apk add --update bash net-tools iproute2 logrotate less rsync util-linux lvm2
WORKDIR /
COPY --from=builder /go/src/github.com/koordinator-sh/koordetector/interference-manager .
ENTRYPOINT ["/interference-manager"]
