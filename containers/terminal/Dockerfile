FROM node:24-trixie-slim

# ttyd provides the in-browser terminal; kubectl execs into the sandbox pod.
# Neither package is in the Debian repos, so fetch static binaries.
ARG TTYD_VERSION=1.7.7
ARG KUBECTL_VERSION=v1.33.1

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

RUN curl -fsSLo /usr/local/bin/ttyd \
      "https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/ttyd.x86_64" && \
    chmod +x /usr/local/bin/ttyd

RUN curl -fsSLO \
      "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

WORKDIR /app

COPY terminal-entrypoint.sh /terminal-entrypoint.sh
RUN chmod +x /terminal-entrypoint.sh

EXPOSE 7681

CMD ["/terminal-entrypoint.sh"]