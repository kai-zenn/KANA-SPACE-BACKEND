# Stage 1: Build
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy go.mod dan go.sum dulu
COPY go.mod go.sum ./
RUN go mod download

# Copy semua source code
COPY . .

# Build binary, target ke cmd/main.go
RUN go build -o main cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates dan timezone data (opsional)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary dari builder
COPY --from=builder /app/main .

# Copy folder configs (kalau ada file config tambahan)
COPY --from=builder /app/internal/configs ./internal/configs

# Buat folder uploads (kosong)
RUN mkdir -p uploads

# Expose port
EXPOSE 9090

# Jalankan binary
CMD ["./main"]
