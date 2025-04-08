ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set working directory
WORKDIR /home/app/services
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd/auth/

FROM debian:bookworm-slim

WORKDIR /home/app/services

COPY --from=builder /home/app/services/auth-service .
COPY deploy/auth/development/config.yaml ./deploy/auth/development/config.yaml

CMD ["./auth-service"]
