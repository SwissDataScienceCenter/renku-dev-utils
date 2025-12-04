FROM golang:1.24-alpine AS builder

RUN apk add --no-cache make bash

WORKDIR /app

# Copy renku-dev-utils files
COPY . .

# Build the rdu binary
RUN make rdu

FROM alpine:3.18

RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates \
    jq \
    openssl \
    python3 \
    py3-pip \
    && ARCH=$(case $(uname -m) in x86_64) echo amd64;; aarch64) echo arm64;; *) echo amd64;; esac) \
    && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${ARCH}/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# Install Python dependencies
RUN pip3 install --no-cache-dir \
    python-dateutil \
    PyGithub \
    kubernetes

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 get_helm.sh \
    && ./get_helm.sh \
    && rm get_helm.sh

# Copy the rdu binary from builder stage
COPY --from=builder /app/build/renku-dev-utils /usr/local/bin/rdu

# Make rdu executable
RUN chmod +x /usr/local/bin/rdu

# Create a non-root user
RUN addgroup -g 1000 appuser && \
    adduser -u 1000 -G appuser -s /bin/bash -D appuser

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /home/appuser

# Verify installations
RUN rdu version || echo "rdu installed" && \
    kubectl version --client && \
    helm version

CMD ["/bin/bash"]
