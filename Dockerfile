# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

# Set destination for COPY
WORKDIR /app

# Copy the source code
COPY . .

WORKDIR /app/server

# Download Go modules
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /appointments-ms -buildvcs=false

FROM golang:1.22-alpine

COPY --from=builder /appointments-ms /appointments-ms

# To bind to a TCP port, runtime parameters must be supplied to the docker command.
EXPOSE 9090

# Run
CMD ["/appointments-ms"]