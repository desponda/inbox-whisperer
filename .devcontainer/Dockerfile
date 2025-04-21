# Custom Devcontainer for Docker, Go, and React
FROM ubuntu:22.04

# Set noninteractive installation
ENV DEBIAN_FRONTEND=noninteractive

# Install core tools for Cascade/Windsurf and dev workflows
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        apt-transport-https \
        ca-certificates \
        curl \
        wget \
        gnupg \
        lsb-release \
        python3 \
        python3-pip \
        build-essential \
        bash \
        tzdata \
        git \
        sudo \
        nodejs \
        npm \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        docker-ce-cli \
        docker-compose-plugin \
    # [Optional] Install Cascade/Windsurf agent or CLI here if required
    # RUN pip install cascade-cli  # Example only, uncomment and modify if needed
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GO_VERSION=1.24.0
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt
ENV PATH=$PATH:/usr/local/go/bin

# Install Node.js (LTS) and npm
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -y nodejs

# Install yarn (optional, for React)
RUN npm install -g yarn

# Install common dev tools
RUN apt-get install -y git sudo

# Clean up
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

# Reset DEBIAN_FRONTEND
ENV DEBIAN_FRONTEND=dialog

# Set default shell
SHELL ["/bin/bash", "-c"]
