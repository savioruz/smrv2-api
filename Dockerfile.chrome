# Stage 1: Build the Go application
FROM golang:1.23-bullseye AS builder
LABEL maintainer="savioruz <jakueenak@gmail.com>"

# Move to working directory (/build).
WORKDIR /build

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container.
COPY . .

# Set necessary environment variables needed for our image and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o smrv2-api ./cmd/app/main.go

# Stage 2: Setup the runtime environment
FROM debian:bullseye-slim

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install necessary runtime dependencies, including Chromium
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    chromium \
    chromium-driver \
    fonts-freefont-ttf \
    libasound2 \
    libx11-6 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxi6 \
    libxtst6 \
    libxrandr2 \
    libxss1 \
    libglib2.0-0 \
    libnss3 \
    libcups2 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libpangocairo-1.0-0 \
    libgtk-3-0 \
    wget \
    xvfb \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Update the CA certificates
RUN update-ca-certificates

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/build/smrv2-api", "/"]

# Set up environment variables for Chromium
ENV CHROME_BIN=/usr/bin/chromium \
    CHROME_PATH=/usr/lib/chromium/ \
    DISPLAY=:99

# Create a symlink for google-chrome (if needed)
RUN ln -s /usr/bin/chromium /usr/bin/google-chrome

# Add a script to start Xvfb before running the main application
RUN echo '#!/bin/bash\nXvfb :99 -ac &\nexec "$@"' > /entrypoint.sh \
    && chmod +x /entrypoint.sh

# Use the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

# Command to run when starting the container.
CMD ["/smrv2-api"]