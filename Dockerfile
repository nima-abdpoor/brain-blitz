ARG GO_IMAGE_VERSION=1.26
FROM golang:${GO_IMAGE_VERSION}-bookworm AS source
WORKDIR /src
ARG GOPROXY=https://proxy.golang.org|direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM source AS check
CMD ["sh", "scripts/check.sh"]

FROM source AS build
ARG SERVICE
RUN test -n "$SERVICE" && CGO_ENABLED=0 go build -trimpath -o /out/service ./cmd/${SERVICE}/

FROM alpine:3.23 AS runtime
WORKDIR /app
RUN addgroup -g 10001 app && adduser -D -u 10001 -G app app && chown app:app /app
COPY --from=build /out/service ./service
COPY --from=source /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chown=10001:10001 infra/deploy ./infra/deploy
COPY --chown=10001:10001 services/user_app/repository/migrations ./services/user_app/repository/migrations
COPY --chown=10001:10001 services/question_app/repository/migrations ./services/question_app/repository/migrations
USER app
ENTRYPOINT ["./service"]
