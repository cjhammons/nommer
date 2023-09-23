# Use the official golang image from the Docker Hub
FROM golang:1.17 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go mod and sum files, and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application for Linux x86
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Use a minimal alpine image
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /app/main /main

# Command to run
CMD ["/main"]
