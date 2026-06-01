# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

ARG CODEX_VERSION=latest
ARG CLAUDE_CODE_VERSION=latest
ARG OPENCODE_VERSION=latest
RUN npm install -g "@openai/codex@${CODEX_VERSION}" \
 && npm install -g "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
 && npm install -g "opencode-ai@${OPENCODE_VERSION}" \
 && command -v codex \
 && command -v claude \
 && command -v opencode
