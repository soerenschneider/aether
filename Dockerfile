FROM rust:1.99.1-alpine3.24 AS build-taskwarrior

# Install dependencies needed to build Taskwarrior
RUN apk add --no-cache \
    build-base \
    git \
    cmake \
    libunwind-dev \
    util-linux-dev

WORKDIR /taskwarrior

# renovate: datasource=github-releases depName=GothenburgBitFactory/taskwarrior
ARG TASKWARRIOR_VERSION=v3.5.0
RUN git clone --depth 1 --branch ${TASKWARRIOR_VERSION} --recurse-submodules --shallow-submodules \
    https://github.com/GothenburgBitFactory/taskwarrior.git .

RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && \
    cmake --build build -j "$(nproc)" && \
    cmake --install build

FROM golang:1.26.10 AS build-aether

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.24.3 AS final

LABEL maintainer="soerenschneider"

RUN addgroup -g 65532 aether && \
    adduser -D -u 65532 -G aether aether

RUN apk add --no-cache \
    tzdata \
    libuuid \
    libunwind \
    libstdc++ \
    libgcc

COPY --from=build-taskwarrior /usr/local/bin/task /usr/bin/task
COPY --from=build-aether /src/aether /aether

USER aether:aether
ENTRYPOINT ["/aether"]
