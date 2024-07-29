# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS base

# Install dependencies only when needed
FROM base AS deps

# Download git
RUN apk add --no-cache git

WORKDIR /app

# Copy protobuf files
COPY ./appointments_protobuf ./appointments_protobuf

# Copy go.mod and go.sum
COPY ./server/go.mod ./server/go.mod
COPY ./server/go.sum ./server/go.sum

WORKDIR /app/server

# Build argument for GitHub token
ARG GITHUB_ACTOR
ARG GITHUB_TOKEN

# Set up git configuration to use token for private repo
RUN git config --global url."https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Download Go modules
RUN go mod download

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps $GOPATH $GOPATH

# Copy the source code
COPY . .

WORKDIR /app/server

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /appointments-ms

# Production image, copy all the files and run
FROM golang:1.22-alpine

COPY --from=builder /appointments-ms /appointments-ms

# To bind to a TCP port, runtime parameters must be supplied to the docker command.
EXPOSE 9090

# Run
CMD ["/appointments-ms"]