# Build stage
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set Go proxy for China
ENV GOPROXY=https://goproxy.cn,direct

# Copy go mod files and download dependencies
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Generate go.sum and build for target platform
ARG TARGETOS
ARG TARGETARCH
RUN go mod tidy && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /app/server .

# Runtime stage
FROM --platform=$TARGETPLATFORM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Set timezone
ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./server"]
