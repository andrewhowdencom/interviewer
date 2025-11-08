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

# Build the application
# Build a dynamically-linked binary for the debian image.
# -o /vox creates the binary at the root of the filesystem, named "vox"
RUN go build -o /vox .

# Start a new, smaller stage from debian:stable
FROM debian:stable

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /vox /vox

# Add custom resolv.conf
COPY resolv.conf /etc/resolv.conf

# Expose the port that the server will listen on
EXPOSE 8080

# Set the binary as the entrypoint for the container and "serve" as the default command
ENTRYPOINT ["/vox"]
CMD ["serve"]
