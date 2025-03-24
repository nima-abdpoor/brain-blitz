ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set working directory
WORKDIR /home/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o match-service ./cmd/match/

FROM debian:bookworm-slim

WORKDIR /home/app

COPY --from=builder /home/app/match-service .
#COPY deploy/match/development/config.yaml ./deploy/match/development/config.yaml

CMD ["./match-service"]