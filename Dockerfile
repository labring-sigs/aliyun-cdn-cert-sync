FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY configs ./configs
COPY internal ./internal

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -tags=clientgo -o /out/cdn-cert-sync ./cmd/cdn-cert-sync

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/cdn-cert-sync /app/cdn-cert-sync
COPY configs/config.example.yaml /app/configs/config.example.yaml

ENTRYPOINT ["/app/cdn-cert-sync"]
CMD ["--config", "/app/configs/config.example.yaml"]
