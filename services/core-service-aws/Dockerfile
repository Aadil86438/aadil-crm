FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o core-service-aws .
EXPOSE 8080
CMD ["./core-service-aws"]
