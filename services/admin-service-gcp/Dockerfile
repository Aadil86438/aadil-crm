FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -p 1 -ldflags="-s -w" -o admin-service-gcp .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/admin-service-gcp .
EXPOSE 8081
CMD ["./admin-service-gcp"]
