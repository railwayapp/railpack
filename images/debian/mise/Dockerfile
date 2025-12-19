FROM rust:1-slim-bookworm AS builder

# Compile mise from source with size optimizations
RUN apt-get update && apt-get install -y \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --branch v2025.12.12 --depth 1 https://github.com/jdx/mise.git /tmp/mise && \
    cd /tmp/mise && \
    # RUSTFLAGS="-C opt-level=z -C link-arg=-s -C codegen-units=1" \
    cargo build --profile release \
    --config debuginfo=0

RUN cp /tmp/mise/target/release/mise /usr/local/bin/
