FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY deploy/update-agent/main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/update-agent main.go

FROM docker/compose:2.29.7
COPY --from=builder /out/update-agent /usr/local/bin/update-agent
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/update-agent"]
