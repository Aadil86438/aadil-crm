FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o admin-service-gcp .
EXPOSE 8081
CMD ["./admin-service-gcp"]
