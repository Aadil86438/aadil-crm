FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY admin-service-gcp-linux ./admin-service-gcp
RUN chmod +x ./admin-service-gcp
EXPOSE 8081
CMD ["./admin-service-gcp"]
