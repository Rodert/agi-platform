FROM golang:1.24-alpine AS builder

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server ./
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/agi-server ./cmd/server

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/agi-server /usr/local/bin/agi-server
RUN mkdir -p /app/uploads && chown -R app:app /app

ENV HTTP_HOST=0.0.0.0 \
    HTTP_PORT=18082

EXPOSE 18082
USER app
CMD ["agi-server"]
