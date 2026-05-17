FROM docker.io/library/golang:1.26.1 AS builder

WORKDIR /src

COPY . /src

RUN mkdir -p /out/bin /out/conf && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/bin/v2ray /src/main/main.go && \
    cp /src/conf/config.json /out/conf/config.json

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

COPY --from=builder /out/bin /app/bin
COPY --from=builder /out/conf /app/conf

ENTRYPOINT ["/app/bin/v2ray", "run", "-config", "/app/conf/config.json"]
