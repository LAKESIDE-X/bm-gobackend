# Step 1: Use the official Go 1.26 image to build the app
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files and install them
COPY go.mod go.sum ./
RUN go mod download

# Copy your source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Step 2: Use a tiny, secure base image to run the app
FROM alpine:latest  
WORKDIR /root/

# Copy the pre-built binary from the previous step
COPY --from=builder /app/main .

# Expose the port (adjust if your app uses something other than 8000)
EXPOSE 8000

# Run the binary
CMD ["./main"]