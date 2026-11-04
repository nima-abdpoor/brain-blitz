ARG GO_IMAGE_NAME=golang
ARG GO_IMAGE_VERSION=1.26
FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}-bookworm AS builder
WORKDIR /src
ARG GOPROXY=https://proxy.golang.org|direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/service ./cmd/game/

FROM alpine:3.23
WORKDIR /app
RUN addgroup -g 10001 app && adduser -D -u 10001 -G app app && chown app:app /app
COPY --from=builder /out/service ./service
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chown=10001:10001 infra/deploy/game/development/config.yaml ./infra/deploy/game/development/config.yaml
USER app
ENTRYPOINT ["./service"]
