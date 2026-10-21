# Base image for the final runtime image created by railpack
FROM debian:trixie-slim@sha256:62bd423aeee912639ca8ee681c2b762cb3e25137e0a92b20be455898a9f034ad

ENV DEBIAN_FRONTEND=noninteractive

# eMudhra fix: ca-certificates >= 20250419+deb13u1.
RUN apt-get update && apt-get install -y \
    ca-certificates \
    pkg-config \
    openssl \
    tzdata \
    locales \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' ca-certificates)" ge '20250419+deb13u1' \
    && sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen \
    && locale-gen \
    && rm -rf /var/lib/apt/lists/*
