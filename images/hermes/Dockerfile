# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

RUN tee /usr/local/bin/hermes >/dev/null <<'EOF' && chmod +x /usr/local/bin/hermes
#!/usr/bin/env bash
echo "Hermes profile image is ready. Install a Hermes-compatible CLI in this image when one is selected."
EOF
