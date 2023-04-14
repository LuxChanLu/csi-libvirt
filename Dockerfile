ARG BUILD_FLAGS=
FROM golang:1.20.3-alpine3.17 as decorate

ENV CGO_ENABLED 0

RUN adduser --disabled-password --gecos "" --home "/app" --shell "/sbin/nologin" --no-create-home --uid "10001" "csi"

WORKDIR /app/csi

COPY ./ ./

RUN go build $BUILD_FLAGS -o ./csi cmd/main.go

FROM scratch

USER csi:csi

ENTRYPOINT ["/app/csi"]

COPY --from=decorate /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=decorate /etc/passwd /etc/passwd
COPY --from=decorate /etc/group /etc/group

COPY --from=decorate /app/csi/csi /app/csi