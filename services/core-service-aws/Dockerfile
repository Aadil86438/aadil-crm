FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o core-service-aws .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/core-service-aws .
COPY migrations/ ./migrations/
EXPOSE 8080
CMD ["./core-service-aws"]
