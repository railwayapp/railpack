# Base image for the final runtime image created by railpack
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    ca-certificates \
    pkg-config \
    openssl \
    && rm -rf /var/lib/apt/lists/*
