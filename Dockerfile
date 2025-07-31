FROM golang:1.21-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG COMMIT_HASH=unknown

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o /app/awsm ./cmd/awsm

# Create a minimal image
FROM alpine:3.18

# Install CA certificates and bash
RUN apk --no-cache add ca-certificates bash

# Copy the binary from the builder stage
COPY --from=builder /app/awsm /usr/local/bin/awsm

# Set entrypoint
ENTRYPOINT ["awsm"]

# Default command
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="awsm"
LABEL org.opencontainers.image.description="AWS CLI Made Awesome"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_TIME}"
LABEL org.opencontainers.image.revision="${COMMIT_HASH}"
LABEL org.opencontainers.image.source="https://github.com/ao/awsm"