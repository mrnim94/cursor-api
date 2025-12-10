# syntax=docker/dockerfile:1.20

ARG GO_VERSION=1.25.1
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# ===== Build stage =====
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/server .

# ===== Runtime stage =====
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl bash xz-utils tar && \
    rm -rf /var/lib/apt/lists/*

# Pre-create log directories expected by the Go logger and make them writable
RUN mkdir -p /log_files/info /log_files/error && chown -R 10001:10001 /log_files

# Non-root user
RUN useradd -m -u 10001 appuser

# Install Cursor CLI for headless usage
USER appuser
ENV PATH="/home/appuser/.local/bin:/usr/local/bin:${PATH}"
RUN bash -c 'curl https://cursor.com/install -fsSL | bash && "$HOME/.local/bin/cursor-agent" --version'

# App binary
USER root
RUN ln -sf /home/appuser/.local/bin/cursor-agent /usr/local/bin/cursor-agent
WORKDIR /app
COPY --chown=10001:10001 --from=builder /out/server /app/server

USER appuser
EXPOSE 1994
# Inject at runtime (K8s/Docker secrets)
ENV CURSOR_API_KEY=""

# Optional: print version on start for quick diagnostics
# HEALTHCHECK can be added if you expose /health in your app
CMD ["/app/server"]
