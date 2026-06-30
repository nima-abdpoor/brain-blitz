ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set working directory
WORKDIR /home/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o game-service ./cmd/game/

FROM debian:bookworm-slim

WORKDIR /home/app/services

COPY --from=builder /home/app/game-service .
COPY infra/deploy/game/development/config.yaml ./deploy/game/development/config.yaml

CMD ["./game-service"]
