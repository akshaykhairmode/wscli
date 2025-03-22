FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./

# Download all dependencies. Caching is leveraged here.
RUN go mod download

# Copy the source from the host to the container
COPY . .

# Build the Go application
ARG GIT_TAG
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.CLIVersion=${GIT_TAG} -w -s" -o /app/wscli main.go

# --- Final Stage ---
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /app/wscli /usr/local/bin/wscli

# Make the binary executable (if necessary, should be done by build stage)
# RUN chmod +x /usr/local/bin/wscli

# Set the entry point for the container
ENTRYPOINT ["wscli"]

# Optionally, define the default command arguments
CMD ["--help"]