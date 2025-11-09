# Use the official Golang image to build the application
# This is a multi-stage Dockerfile, as it uses a build image and a final, smaller image
FROM golang:latest AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application's source code
COPY . .

# Build a statically-linked binary
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /vox .

# Start a new, smaller stage from scratch
FROM scratch

# Copy the CA certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --from=builder /vox /vox

# Expose the port that the server will listen on
EXPOSE 8080

# Set the binary as the entrypoint for the container and "serve" as the default command
ENTRYPOINT ["/vox"]
CMD ["serve"]
