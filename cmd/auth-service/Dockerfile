FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service ./cmd/auth-service/

FROM alpine:3.19

RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

WORKDIR /home/app

COPY --from=builder /app/auth-service .
COPY --from=builder /app/config.yaml.example ./config.yaml

USER app

EXPOSE 8080

ENTRYPOINT ["./auth-service"]
CMD ["--config", "config.yaml"]
