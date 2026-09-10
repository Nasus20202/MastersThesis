FROM ubuntu:26.04@sha256:513c074113a871b51a8d16ab445c88779d6452d937a164fb5cc479f32668a41d

ARG KUBECTL_VERSION=v1.37.0
ARG KUBECTL_SHA256=6129359f4e1f3848a5572ccb0b26cf28b8ca08cef38c95a765b2f64a2c961a2f

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install --no-install-recommends --yes \
        bash \
        ca-certificates \
        coreutils \
        curl \
        jq \
        nodejs \
        python3 \
    && curl --fail --location --retry 3 \
        "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
        --output /usr/local/bin/kubectl \
    && echo "${KUBECTL_SHA256}  /usr/local/bin/kubectl" | sha256sum --check --strict \
    && chmod 0755 /usr/local/bin/kubectl \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --home-dir /home/benchmark --shell /bin/bash benchmark \
    && install --directory --owner=benchmark --group=benchmark /home/benchmark/.kube /workspace

ENV HOME=/home/benchmark

WORKDIR /workspace
USER benchmark
