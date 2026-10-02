# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.27-alpine AS builder

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w -X github.com/kirasync2748/github-widget/internal/version.Version=${VERSION}" \
      -o /server ./cmd/server

# ---- Runtime stage ----
FROM scratch

COPY --from=builder /server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV PORT=3000
EXPOSE 3000

USER 65532:65532

ENTRYPOINT ["/server"]
