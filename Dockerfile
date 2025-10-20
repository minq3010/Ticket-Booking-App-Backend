FROM golang:alpine

# Set working directory
WORKDIR /src/app

# Install git, curl, etc.
RUN apk add --no-cache git curl

# Install air for live reload
# RUN go install github.com/cosmtrek/air@v1.40.4

# Copy go.mod and download deps early
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build cleanup (optional)
RUN go mod tidy

RUN go build -o main .

# Expose port (for Fiber or similar)
EXPOSE 8080

# Default command: use air to auto-reload
CMD ["./main"]
