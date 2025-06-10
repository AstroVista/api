# Use the official Go image for the build phase (compatible with go.mod)
FROM golang:1.24.3 AS builder

# Define the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Copy Go module files and download dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy all application source code
COPY . .

# Build the Go executable (statically linked with no system dependencies)
# This creates an optimized and standalone executable
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /goapi .

# Use a distroless base image that already includes CA certificates and timezone data
FROM gcr.io/distroless/cc-debian11

# Add CA certificates (already included in the base image, but kept for documentation)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the built executable from the previous stage to the final image
COPY --from=builder /goapi /goapi

# Copy the localization files needed for i18n
COPY --from=builder /app/i18n/locales /i18n/locales

# Copy the .env file for configuration
COPY --from=builder /app/.env /.env

# Define the working directory
WORKDIR /

# Expose the port your application listens on
EXPOSE 8080

# Command to run the application when the container starts
CMD ["/goapi"]