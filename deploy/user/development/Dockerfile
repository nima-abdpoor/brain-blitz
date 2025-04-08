ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set working directory
WORKDIR /home/app/services
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o user-service ./cmd/user/

FROM debian:bookworm-slim

WORKDIR /home/app/services

COPY --from=builder /home/app/services/user-service .
COPY deploy/user/development/config.yaml ./deploy/user/development/config.yaml
COPY services/user_app/repository/migrations/* ./services/user_app/repository/migrations/

CMD ["./user-service"]
