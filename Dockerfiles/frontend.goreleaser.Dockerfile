# syntax=docker.io/docker/dockerfile:1
FROM alpine

COPY railpack /

ENTRYPOINT ["/railpack", "frontend"]
