FROM golang:1.19 as builder

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o external-scaler main.go


FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /
COPY --from=builder /src/external-scaler .
ENTRYPOINT ["/external-scaler"]
