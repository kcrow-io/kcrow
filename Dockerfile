# Copyright 2023 Authors of kcrow
# SPDX-License-Identifier: Apache-2.0

ARG BASE_IMAGE=docker.io/library/busybox:1.36.1

FROM ${BASE_IMAGE}

# TARGETOS is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETOS
# TARGETARCH is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETARCH

COPY bin/*   /usr/bin/
CMD ["/usr/bin/daemon daemon"]
