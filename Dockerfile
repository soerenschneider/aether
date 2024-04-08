FROM golang:1.22.2 AS build

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
ENV CGO_ENABLED=0
RUN go mod download
RUN make build

FROM alpine:3.19.1 AS final

LABEL maintainer="soerenschneider"

RUN addgroup -g 65532 aether && \
    adduser -D -u 65532 -G aether aether

RUN apk add --no-cache task

COPY --from=build /src/aether /aether
USER aether:aether

ENTRYPOINT ["/aether"]
