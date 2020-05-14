# syntax=docker/dockerfile:experimental

# Build vctui
FROM golang:1.14-alpine as dev
# add make / gcc (CGO_ENABLED)
RUN apk add --no-cache git ca-certificates make gcc linux-headers musl-dev
COPY . /go/src/github.com/thebsdbox/vctui
WORKDIR /go/src/github.com/thebsdbox/vctui
ENV GO111MODULE=on
RUN --mount=type=cache,sharing=locked,id=gomod,target=/go/pkg/mod/cache \
    --mount=type=cache,sharing=locked,id=goroot,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux make build

FROM scratch
COPY --from=dev /go/src/github.com/thebsdbox/vctui/vctui /bin/vctui
ENTRYPOINT ["/bin/vctui"]