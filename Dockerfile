FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o imgbed .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/imgbed .
COPY --from=builder /app/static ./static

EXPOSE 8080
VOLUME ["/app/uploads", "/app/data"]

ENV UPLOAD_DIR=/app/uploads
ENV DB_PATH=/app/data/imgbed.db

ENTRYPOINT ["./imgbed"]
