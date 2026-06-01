# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

ARG CLAUDE_CODE_VERSION=latest
RUN npm install -g "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
 && command -v claude
