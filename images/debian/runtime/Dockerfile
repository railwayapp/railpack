# Base image for the final runtime image created by railpack
FROM debian:bookworm-slim@sha256:7ff6ba31527b9e336162fa061746c366c4b2032f4d2cff2078bea43f78efe3f8

ENV DEBIAN_FRONTEND=noninteractive

# eMudhra fix: ca-certificates >= 20250419~deb12u2.
RUN apt-get update && apt-get install -y \
    ca-certificates \
    pkg-config \
    openssl \
    tzdata \
    locales \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' ca-certificates)" ge '20250419~deb12u2' \
    && sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen \
    && locale-gen \
    && rm -rf /var/lib/apt/lists/*
