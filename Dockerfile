# Build: lpac
FROM alpine:3.20 AS lpac-builder

WORKDIR /app

RUN apk add --no-cache git gcc cmake make musl-dev curl-dev

COPY . .

RUN set -eux \
    && cd lpac \
    && cmake . -DLPAC_WITH_APDU_PCSC=off -DLPAC_WITH_APDU_AT=off \
    && make -j$(nproc)

# Build: estkme-cloud
FROM golang:1.22-alpine AS estkme-cloud-builder

WORKDIR /app

ARG VERSION

COPY . .

RUN set -eux \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-w -s -X main.Version=${VERSION}" -o estkme-cloud main.go

# Production
FROM alpine:3.20 AS production

WORKDIR /app

COPY --from=lpac-builder /app/lpac/output/lpac /app/lpac
COPY --from=estkme-cloud-builder /app/estkme-cloud /app/estkme-cloud

RUN set -eux \
    && apk add --no-cache libcurl

EXPOSE 1888

CMD ["/app/estkme-cloud", "--dont-download", "--dir=/app"]
