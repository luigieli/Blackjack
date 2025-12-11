# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod ./
# COPY go.sum ./ # No go.sum yet, but good practice usually
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o blackjack-api .

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/blackjack-api .

# Copy frontend assets for production
# Note: In development with docker-compose, this might be shadowed by the volume mount, which is fine.
# We need to copy from the build context.
# However, the build context in docker-compose is ./backend.
# So we can't easily copy ../frontend unless we change context or copy it in Stage 1.
# But Dockerfile context is backend/.
# A common pattern is to have the build context at root.
# Since we can't change the build context easily without changing docker-compose...
# AND the user specified backend/Dockerfile.
# I will leave this as is but add a comment or try to accommodate.
# Actually, I can't copy ../frontend if context is backend.
# So for this specific setup, we rely on the volume mount or assume frontend files are copied to backend/frontend before build in a CI pipeline.
# BUT, I can create a 'frontend' directory here to avoid crash if not mounted.
RUN mkdir -p frontend

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./blackjack-api"]
