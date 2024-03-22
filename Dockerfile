FROM golang:1.22.1-bookworm as builder

ENV VERSION=0.0.2-alpha

WORKDIR /app

COPY . .

RUN set -ex \
    && go mod download \
    && go build -trimpath -ldflags="-w -s -X main.version=${VERSION}" -o estkme-rlpa-server main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/estkme-rlpa-server /app/estkme-rlpa-server

RUN set -ex \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates libpcsclite1 libcurl4 \
    && rm -rf /var/lib/apt/lists/*

RUN chmod +x /app/estkme-rlpa-server

EXPOSE 1888

CMD ["/app/estkme-rlpa-server"]
