FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o running-tracker ./cmd/server

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/running-tracker .
COPY --from=builder /app/openapi.yaml .
EXPOSE 8080
ENTRYPOINT ["./running-tracker"]
