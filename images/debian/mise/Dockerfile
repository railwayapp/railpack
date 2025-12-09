# syntax=docker.io/docker/dockerfile:1
FROM rust:1-slim-bookworm AS builder

# Compile mise from source with size optimizations
RUN apt-get update && apt-get install -y \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/jdx/mise.git /tmp/mise && \
    cd /tmp/mise && \
    RUSTFLAGS="-C opt-level=z -C link-arg=-s -C codegen-units=1" \
    cargo build --profile serious \
    --config debuginfo=0

RUN cp /tmp/mise/target/serious/mise /usr/local/bin/
