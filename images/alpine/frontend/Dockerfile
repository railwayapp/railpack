# Not used by goreleaser, for local development
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# TODO we should really pull from github releases here instead of recompiling, but allow for a local development build for builtctl testing
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/railpack cmd/cli/main.go

FROM alpine

COPY --from=builder /usr/bin/railpack /usr/bin/railpack
ENTRYPOINT ["/usr/bin/railpack", "frontend"]
