FROM golang:1.23-bullseye AS builder
LABEL maintainer="savioruz <jakueenak@gmail.com>"

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o smrv2-api ./cmd/app/main.go

FROM chromedp/headless-shell:latest

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates dumb-init \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

WORKDIR /app

COPY --from=builder /build/smrv2-api /app/
COPY .env /app/

RUN chmod +x /app/smrv2-api

EXPOSE 3000

ENTRYPOINT ["dumb-init", "--"]

CMD ["./smrv2-api"]
