# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /workspace

# Download dependencies first (separate layer for cache efficiency)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /workspace/server ./cmd/server/

# ─── Stage 2: Runtime (distroless — no shell, no package manager) ─────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /workspace/server /app/server

# nonroot user (UID 65532) is built into distroless:nonroot
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/server"]
