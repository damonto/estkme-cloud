FROM golang:1.22.1-alpine as builder

WORKDIR /app

COPY . .

RUN set -ex \
    && go mod download \
    && go build -trimpath -ldflags="-w -s" -o estkme-cloud main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/estkme-cloud /app/estkme-cloud

RUN set -ex \
    && apk add --no-cache gcompat ca-certificates pcsc-lite-libs libcurl \
    && update-ca-certificates \
    && chmod +x /app/estkme-cloud

EXPOSE 1888

CMD ["/app/estkme-cloud"]
