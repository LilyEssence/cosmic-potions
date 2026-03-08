# ── Stage 1: Build ───────────────────────────────────────────────────
# GO CONCEPT: Multi-Stage Docker Builds
# Go compiles to a single static binary — no runtime, no node_modules, no
# interpreter. A multi-stage build exploits this: Stage 1 has the full Go
# toolchain (~800MB) to compile. Stage 2 copies ONLY the binary (~15MB)
# into a minimal Alpine image (~5MB). Final image: ~20MB total.
#
# This is a huge advantage over Node.js deploys where you ship the entire
# runtime + node_modules. Go's single-binary output is one of its best
# deployment stories.

FROM golang:1.24-alpine AS builder

# Install certificates for HTTPS calls (needed at runtime too).
RUN apk add --no-cache ca-certificates

WORKDIR /build

# Copy dependency files first — Docker caches this layer so `go mod download`
# only re-runs when go.mod/go.sum change, not on every code edit.
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build.
COPY . .

# GO CONCEPT: CGO_ENABLED=0
# modernc.org/sqlite is pure Go (no C code), so we can disable CGO entirely.
# This produces a fully static binary that runs on any Linux — no libc needed.
# Without this, the binary might dynamically link to glibc, which doesn't
# exist in Alpine (Alpine uses musl). CGO_ENABLED=0 avoids that mismatch.
#
# GO CONCEPT: -ldflags="-s -w"
# Strips debug symbols and DWARF info from the binary, reducing size by ~30%.
# We don't need debugger support in production.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cosmic-potions ./cmd/server

# ── Stage 2: Runtime ─────────────────────────────────────────────────
# Alpine is ~5MB. We only need the binary, migrations folder, and CA certs.

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage.
COPY --from=builder /build/cosmic-potions .

# Copy migrations — needed at startup to initialize/update the SQLite schema.
COPY --from=builder /build/migrations ./migrations

# Railway injects PORT at runtime. Default to 8080 for local Docker testing.
EXPOSE 8080

# GO CONCEPT: Entrypoint vs CMD
# ENTRYPOINT sets the executable; CMD provides default arguments. Using just
# CMD means `docker run <image>` runs this command, but you can override it.
# For a single-binary service, either works. CMD is simpler.
CMD ["./cosmic-potions"]
