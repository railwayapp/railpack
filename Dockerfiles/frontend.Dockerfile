FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/railpack cmd/cli/main.go

FROM alpine

COPY --from=builder /usr/bin/railpack /usr/bin/railpack
ENTRYPOINT ["/usr/bin/railpack", "frontend"]
