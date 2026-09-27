# ==========================================
# STAGE 1: Build the Go binary
# ==========================================
FROM golang:1.27.1-alpine AS builder

# Install build tools (git, ca-certificates)
RUN apk add --no-cache git ca-certificates

# Create user used in final stage
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Cache dependencies by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile a statically linked binary with CGO disabled
RUN CGO_ENABLED=0 GOOS=linux go build \
   -ldflags="-s -w" \
   -o pdfserver .

# ==========================================
# STAGE 2: Minimal Runtime Container
# ==========================================
FROM scratch AS final

# Copy CA Certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy users from build stage
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Run as a non-root user for security
USER appuser

WORKDIR /app

# Copy the compiled binary and templates from the builder stage
COPY --from=builder /app/pdfserver .

# Expose HTTP port
EXPOSE 8080

# Run the binary
CMD ["./pdfserver"]