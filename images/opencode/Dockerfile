# syntax=docker/dockerfile:1.7
ARG BASE_IMAGE=ghcr.io/infrasecture/myharness-base:latest
FROM ${BASE_IMAGE}

ARG OPENCODE_VERSION=latest
RUN npm install -g "opencode-ai@${OPENCODE_VERSION}" \
 && command -v opencode
