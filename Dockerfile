# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod download (some modules need it)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o predictor .

# Stage 2: Minimal runtime image
FROM alpine:latest

# ca-certificates required for HTTPS calls to external APIs
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/predictor .
COPY --from=builder /app/mock ./mock
COPY --from=builder /app/data ./data

# Optional FM CSV volume mount point
RUN mkdir -p /root/data

EXPOSE 8080

CMD ["./predictor", "serve"]
