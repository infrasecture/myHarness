# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

ARG CODEX_VERSION=latest
RUN npm install -g "@openai/codex@${CODEX_VERSION}" \
 && command -v codex \
 && codex --version

RUN mkdir -p /etc/myharness/profile.d && tee /etc/myharness/profile.d/codex.sh >/dev/null <<'EOF'
export CODEX_HOME="${CODEX_HOME:-/home/myharness/.codex}"
EOF
