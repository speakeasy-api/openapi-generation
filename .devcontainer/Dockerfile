FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Install necessary packages
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    git \
    make \
    build-essential \
    ca-certificates \
    openjdk-11-jdk \
    python3 \
    python3-pip \
    ruby \
    ruby-dev \
    software-properties-common \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install Go
COPY --from=golang:1.22-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"

# Set up Go workspace
RUN mkdir -p $GOPATH/src $GOPATH/bin && chmod -R 777 $GOPATH

# Set up GOPRIVATE
# Note: credentials come from the parent host machine. initializeCommand should make sure they're available. 
ENV GOPRIVATE="github.com/speakeasy-api/*"

ENV DOTNET_ROOT=/usr/lib/dotnet

# Install .NET 5 via binary (Note that .NET 5 is required for the Unity target)
RUN curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin -Channel 5.0 -InstallDir ${DOTNET_ROOT} \
    && ln -s ${DOTNET_ROOT}/dotnet /usr/local/bin/dotnet

# Install .NET 6 via binary
RUN curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin -Channel 6.0 -InstallDir ${DOTNET_ROOT}

# Install .NET 8 via binary and write list of available sdks to log
RUN curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin -Channel 8.0 -InstallDir ${DOTNET_ROOT} \
    && /usr/local/bin/dotnet --list-sdks

# Get a version of libssl installed that .NET 5 can understand
# Unfortunately arm architecture binaries are in a different repo
RUN arch=$(dpkg --print-architecture) \
    && (if [ $arch = "arm64" ]; then \
    url="http://ports.ubuntu.com/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_$arch.deb"; else \
    url="http://security.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_$arch.deb";\
    fi \
    && wget -O libssl.deb $url) \
    && dpkg -i libssl.deb \
    && rm libssl.deb

RUN apt install libssl-dev

# Install Node.js v18
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs

## Install python v2 stuff
RUN curl -sSL https://install.python-poetry.org | python3 - 

# Enable corepack for pnpm
RUN corepack enable

# Install speakeasy CLI
RUN curl -fsSL https://raw.githubusercontent.com/speakeasy-api/speakeasy/main/install.sh | sh

# Set environment variables
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:/usr/bin:/usr/local/bin:$PATH

