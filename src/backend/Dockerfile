# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application statically linked (optional, good for scratch/distroless)
# Set CGO_ENABLED=0 for static build without C dependencies
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/main .

# Stage 2: Create the final lightweight image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main

# Create uploads directory and set permissions
RUN mkdir -p /app/uploads && \
    chown -R 1000:1000 /app/uploads && \
    chmod 755 /app/uploads

# Set the user to non-root for security
RUN addgroup -S -g 1000 appgroup && \
    adduser -S -u 1000 -G appgroup appuser
USER appuser

# Expose the port the application will run on (defined by ENV in docker-compose)
# This is informational, actual port mapping is done in docker-compose
EXPOSE 8080

# Command to run the application
# The PORT env variable will be passed from docker-compose
CMD ["/app/main"]
