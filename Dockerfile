FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bzzz .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bzzz .

# Copy secrets directory for GitHub token access
VOLUME ["/secrets"]

CMD ["./bzzz"]