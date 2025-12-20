# This image is used to build a custom mise binary which is then included in the railpack-builder image
FROM rust:1-slim-bookworm AS builder

ARG MISE_VERSION

RUN apt-get update && apt-get install -y \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --branch v${MISE_VERSION} --depth 1 https://github.com/jdx/mise.git /tmp/mise && \
    cd /tmp/mise && \
    RUSTFLAGS="-C opt-level=z -C link-arg=-s -C codegen-units=1" \
    cargo build --profile serious \
    --config debuginfo=0

RUN cp /tmp/mise/target/serious/mise /usr/local/bin/
