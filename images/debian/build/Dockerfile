FROM buildpack-deps:bookworm-scm

ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
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
    locales \
    && rm -rf /var/lib/apt/lists/*

# Mise config
ENV MISE_INSTALL_PATH=/usr/local/bin/mise \
    MISE_DATA_DIR=/mise \
    MISE_CONFIG_DIR=/mise \
    MISE_CACHE_DIR=/mise/cache \
    PATH=/mise/shims:$PATH

# Copy mise binary from builder
COPY --from=ghcr.io/railwayapp/railpack-mise:latest /usr/local/bin/mise /usr/local/bin/
RUN chmod +x /usr/local/bin/mise

# Make sure the cache directory exists
RUN mkdir -p /root/.cache
