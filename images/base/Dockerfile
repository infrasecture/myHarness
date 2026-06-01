# syntax=docker/dockerfile:1.7
FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

ARG DEBIAN_FRONTEND=noninteractive
SHELL ["/bin/bash", "-lc"]
USER root

ENV NPM_CONFIG_PREFIX=/usr/local
ENV CARGO_HOME=/usr/local/cargo
ENV PATH=/usr/local/cargo/bin:${PATH}
ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8
ENV EDITOR=vi
ENV TERM=xterm-256color
ENV MYHARNESS_HOME=/home/myharness

RUN apt-get update && apt-get install -y \
    byobu tmux sudo util-linux \
    git git-lfs gh \
    ripgrep fd-find curl wget jq yq fzf bat tree htop \
    zip unzip p7zip-full rsync openssh-client \
    lsof strace gdb iproute2 iputils-ping dnsutils net-tools socat \
    ca-certificates locales man-db manpages manpages-dev bash-completion shellcheck \
    build-essential pkg-config cmake make gcc g++ clang clang-format clang-tidy \
    ninja-build meson autoconf automake libtool \
    nodejs npm python3 python3-dev python3-pip python3-venv python-is-python3 \
    rustc cargo vim \
 && locale-gen en_US.UTF-8 \
 && update-locale LANG=en_US.UTF-8 \
 && mkdir -p /usr/local/cargo/bin /workspace /home/myharness /etc/myharness \
 && (ln -sf "$(command -v fdfind)" /usr/local/bin/fd 2>/dev/null || true) \
 && (ln -sf "$(command -v batcat)" /usr/local/bin/bat 2>/dev/null || true) \
 && npm config set fund false \
 && npm config set audit false \
 && npm config set update-notifier false \
 && npm config set prefix "${NPM_CONFIG_PREFIX}" \
 && sed -i '/^#if ! shopt -oq posix; then$/,/^#fi$/ s/^#//' /etc/bash.bashrc \
 && (rm -rf /var/lib/apt/lists/* /tmp/* /tmp/.[a-zA-Z0-9]* 2>/dev/null || true)

COPY runtime/myharness-init.sh /usr/local/bin/myharness-init.sh
COPY runtime/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY runtime/session-banner.txt /etc/myharness/session-banner.txt
RUN chmod +x /usr/local/bin/myharness-init.sh /usr/local/bin/entrypoint.sh

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD []
