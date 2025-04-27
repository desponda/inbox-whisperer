# Custom Devcontainer for Docker, Go, and React
FROM golang:1.24.2

# Set noninteractive installation
ENV DEBIAN_FRONTEND=noninteractive

# Install necessary packages and Docker CLI
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    build-essential \
    bash \
    ca-certificates \
    tzdata \
    gnupg \
    lsb-release \
    binfmt-support \
    qemu-user-static && \
    install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable" > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends docker-ce-cli && \
    rm -rf /var/lib/apt/lists/*

# Install Docker Buildx (latest stable release)
RUN mkdir -p ~/.docker/cli-plugins/ \
    && curl -sSL https://github.com/docker/buildx/releases/latest/download/buildx-linux-amd64 -o ~/.docker/cli-plugins/docker-buildx \
    && chmod +x ~/.docker/cli-plugins/docker-buildx \
    && docker buildx version || true

# Install Node.js 22.x and npm from NodeSource
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs

# Install Go tools and golang-migrate (v4.18.2) and gomock/mockgen
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/golang/mock/mockgen@latest && \
    curl -fL https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz -o migrate.tar.gz && \
    tar -xvzf migrate.tar.gz && \
    chmod +x migrate && \
    mv migrate /usr/local/bin/migrate && \
    rm migrate.tar.gz

# Set up environment
ENV CGO_ENABLED=1 \
    GOOS=linux \
    PATH=$PATH:/go/bin \
    NODE_ENV=development
