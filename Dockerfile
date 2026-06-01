# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

ARG CODEX_VERSION=latest
RUN npm install -g "@openai/codex@${CODEX_VERSION}" \
 && command -v codex \
 && codex --version
