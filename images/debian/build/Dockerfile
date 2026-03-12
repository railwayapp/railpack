# Image used by railpack during the build process, but not as the base image of the final runtime image
FROM buildpack-deps:bookworm-scm

ARG MISE_VERSION
RUN if [ -z "$MISE_VERSION" ] || [ -z "$(echo $MISE_VERSION | tr -d ' ')" ]; then \
  echo "ERROR: MISE_VERSION build argument is required and cannot be empty!" >&2; \
  echo "Usage: docker build --build-arg MISE_VERSION=yourversion ..." >&2; \
  exit 1; \
fi

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    # Build tools
    build-essential \
    cmake \
    make \
    pkg-config \
    autoconf \
    automake \
    libtool \
    # Version control
    git \
    # SSL and encryption
    libssl-dev \
    openssl \
    gnupg \
    ca-certificates \
    # Compression libraries
    zlib1g-dev \
    libbz2-dev \
    liblzma-dev \
    libzstd-dev \
    # Python dependencies
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    # Image processing
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    # Database clients
    default-libmysqlclient-dev \
    libpq-dev \
    # Additional utilities
    jq \
    curl \
    wget \
    zip \
    unzip \
    ssh \
    rsync \
    locales && \
    rm -rf /var/lib/apt/lists/*


# MISE_VERSION do *not* need to persist beyond the build process, but it's a helpful artifact to users for debugging purposes
# MISE_SELF_UPDATE_AVAILABLE is undocumented: https://github.com/iloveitaly/mise/blob/bbe46dd8e24c6b18a177806836c2e951ec3a5c5c/src/env.rs#L238
ENV MISE_INSTALL_PATH=/usr/local/bin/mise \
    MISE_VERSION=${MISE_VERSION} \
    MISE_SELF_UPDATE_AVAILABLE=false \
    MISE_DATA_DIR=/mise \
    MISE_CONFIG_DIR=/mise \
    MISE_CACHE_DIR=/mise/cache \
    PATH=/mise/shims:$PATH

# https://mise.jdx.dev/installing-mise.html
RUN gpg --keyserver hkps://keys.openpgp.org --recv-keys 24853EC9F655CE80B48E6C3A8B81C9D17413A06D && \
  curl https://mise.jdx.dev/install.sh.sig | gpg --decrypt > install_mise.sh && \
  MISE_DEBUG=true sh ./install_mise.sh && \
  rm install_mise.sh

# Make sure the cache directory exists
RUN mkdir -p /root/.cache
