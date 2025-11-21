FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o prmanager ./cmd/pr-reviewer-service

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/prmanager /app/prmanager
COPY config/prod.yaml /config/prod.yaml

ENV CONFIG_PATH=/config/prod.yaml

EXPOSE 8080

ENTRYPOINT ["/app/prmanager"]
