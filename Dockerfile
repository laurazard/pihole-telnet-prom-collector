# Copied (and modified) from https://github.com/docker/cli.
# All credits to the original authors.

ARG ALPINE_VERSION=3.20

ARG GOVERSIONINFO_VERSION=v1.4.1
ARG GOTESTSUM_VERSION=v1.10.0

ARG GO_VERSION=1.23.4
ARG XX_VERSION=1.5.0

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build-base-alpine
ENV GOTOOLCHAIN=local
COPY --link --from=xx / /
RUN apk add --no-cache bash clang lld llvm file git
WORKDIR /go/src/github.com/laurazard/pihole-telnet-prom-collector

FROM build-base-alpine AS build-alpine
ARG TARGETPLATFORM
# gcc is installed for libgcc only
RUN xx-apk add --no-cache musl-dev gcc

FROM build-base-alpine AS goversioninfo
ARG GOVERSIONINFO_VERSION
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOBIN=/out GO111MODULE=on CGO_ENABLED=0 go install "github.com/josephspurrier/goversioninfo/cmd/goversioninfo@${GOVERSIONINFO_VERSION}"

FROM build-base-alpine AS gotestsum
ARG GOTESTSUM_VERSION
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOBIN=/out GO111MODULE=on CGO_ENABLED=0 go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}" \
    && /out/gotestsum --version

FROM build-alpine AS build
# CGO_ENABLED manually sets if cgo is used
ARG CGO_ENABLED
# VERSION sets the version for the produced binary
ARG VERSION
COPY --link --from=goversioninfo /out/goversioninfo /usr/bin/goversioninfo
# Copy golang dependency manifests
COPY go.mod .
COPY go.sum .
# Cache the downloaded dependency in the layer.
RUN go mod download
RUN --mount=type=bind,target=.,ro \
    --mount=type=cache,target=/root/.cache \
    # override the default behavior of go with xx-go
    xx-go --wrap && \
    # TARGET=/out ./scripts/build/binary && \
    CGO_ENABLED=0 go build -o /out/pi-collector -ldflags="-extldflags -static" ./cmd/pi-collector && \
    xx-verify --static /out/pi-collector

FROM build-alpine AS test
COPY --link --from=gotestsum /out/gotestsum /usr/bin/gotestsum
ENV GO111MODULE=auto
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    gotestsum -- -coverprofile=/tmp/coverage.txt $(go list ./... | grep -vE '/vendor/|/e2e/')

FROM scratch AS bin-image-linux
COPY --from=build /out/docker /docker
FROM scratch AS bin-image-darwin
COPY --from=build /out/docker /docker
FROM scratch AS bin-image-windows
COPY --from=build /out/docker /docker.exe

FROM bin-image-${TARGETOS} AS bin-image

FROM scratch AS binary
COPY --from=build /out .

FROM --platform=$BUILDPLATFORM alpine:${ALPINE_VERSION} AS ssh-tunnel-collector
COPY --from=build /out .
COPY ./hack/docker/entrypoint.sh /entrypoint.sh 
RUN apk update
RUN apk add openssh
ENTRYPOINT [ "/entrypoint.sh" ]
