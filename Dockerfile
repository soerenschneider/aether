FROM rust:1.84.1-alpine3.21 AS build-taskwarrior

# Install dependencies needed to build Taskwarrior
RUN apk add --no-cache \
    build-base \
    git \
    cmake \
    libtool \
    autoconf \
    pkgconfig \
    sqlite-dev \
    ncurses-dev \
    libunwind-dev \
    util-linux-dev \
    zlib-dev

# Set the working directory
WORKDIR /taskwarrior

# Clone the Taskwarrior repository
RUN git clone https://github.com/GothenburgBitFactory/taskwarrior.git .

# Build Taskwarrior using cmake
RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release .
RUN cmake --build build -j 8
RUN cmake --install build
RUN task --version

FROM golang:1.24.0 AS build-aether

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
ENV CGO_ENABLED=0
RUN go mod download
RUN make build

FROM alpine:3.21.2 AS final

LABEL maintainer="soerenschneider"

RUN addgroup -g 65532 aether && \
    adduser -D -u 65532 -G aether aether

RUN apk add --no-cache \
    tzdata \
    sqlite-dev \
    libunwind-dev \
    util-linux-dev \
    libstdc++

COPY --from=build-taskwarrior /usr/local/bin/task /usr/bin/task
COPY --from=build-aether /src/aether /aether

USER aether:aether
ENTRYPOINT ["/aether"]
