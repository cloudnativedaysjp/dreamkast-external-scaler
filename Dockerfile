# syntax=docker/dockerfile:1.4
### builder ###
FROM golang:1.20 as builder

WORKDIR /src

# Copy the Go Modules
COPY --link go.mod go.mod
COPY --link go.sum go.sum
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download
COPY --link . .
# Build
ARG GOOS=linux
ARG GOARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "\
  -X github.com/cloudnativedaysjp/seaman/version.Version=${APP_VERSION} \
  -X github.com/cloudnativedaysjp/seaman/version.Commit=${APP_COMMIT} \
  -s -w \
  " -trimpath -a -o external-scaler .

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /
COPY --link --from=builder /src/external-scaler .
ENTRYPOINT ["/external-scaler"]
