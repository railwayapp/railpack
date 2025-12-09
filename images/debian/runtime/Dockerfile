# syntax=docker.io/docker/dockerfile:1
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    ca-certificates \
    pkg-config \
    openssl \
    && rm -rf /var/lib/apt/lists/*
