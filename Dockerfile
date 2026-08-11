FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/forum .

FROM debian:bookworm-slim

RUN groupadd --system forum \
    && useradd --system --gid forum --home-dir /app --no-create-home forum \
    && mkdir -p /app/database \
    && chown -R forum:forum /app

WORKDIR /app

COPY --from=builder /out/forum /app/forum
COPY --from=builder /src/static /app/static

ENV FORUM_HTTP_ADDRESS=:8080 \
    FORUM_DATABASE_PATH=/app/database/forum.db \
    FORUM_STATIC_PATH=/app/static \
    FORUM_ENV=development \
    FORUM_WS_ORIGINS=http://localhost:8080,http://127.0.0.1:8080

VOLUME ["/app/database"]
EXPOSE 8080

USER forum
ENTRYPOINT ["/app/forum"]
