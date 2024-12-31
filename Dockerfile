# Copyright 2023 Authors of kcrow
# SPDX-License-Identifier: Apache-2.0

ARG BASE_IMAGE=docker.io/library/busybox:1.36.1

FROM ${BASE_IMAGE}

ARG GIT_COMMIT_VERSION
ENV GIT_COMMIT_VERSION=${GIT_COMMIT_VERSION}
ARG GIT_COMMIT_TIME
ENV GIT_COMMIT_TIME=${GIT_COMMIT_TIME}
ARG VERSION
ENV VERSION=${VERSION}

COPY bin/*   /usr/bin/
CMD ["/usr/bin/daemon daemon"]
