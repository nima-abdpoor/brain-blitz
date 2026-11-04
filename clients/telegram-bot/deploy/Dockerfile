ARG GO_IMAGE_NAME=golang
ARG GO_IMAGE_VERSION=1.26-bookworm

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set working directory
WORKDIR /home/app
ARG GOPROXY=https://proxy.golang.org|direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o telegram-bot ./cmd/bot/

FROM debian:bookworm-slim

WORKDIR /home/app/

COPY --from=builder /home/app/telegram-bot .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY configs/development/config.yaml ./configs/development/config.yaml

CMD ["./telegram-bot"]
