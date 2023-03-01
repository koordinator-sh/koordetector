FROM golang:1.17 as builder
WORKDIR /go/src/github.com/koordinator-sh/koordetector

COPY ../go.mod go.mod
COPY ../go.sum go.sum

RUN go mod download

# Copy the go source
COPY apis/ apis/
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN GOOS=linux GOARCH=amd64 go build -a -o interference-manager cmd/interference-manager/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/src/github.com/koordinator-sh/koordetector/interference-manager .
ENTRYPOINT ["/interference-manager"]
