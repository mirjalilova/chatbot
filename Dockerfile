# Stage 1: Go build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags migrate -o chatbot ./cmd

# Stage 2: Prepare runtime files in alpine
FROM alpine:latest AS alpine-stage

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/chatbot /app/chatbot
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/internal/controller/http/casbin/model.conf ./internal/controller/http/casbin/
COPY --from=builder /app/internal/controller/http/casbin/policy.csv ./internal/controller/http/casbin/

# Stage 3: Final runtime image with wkhtmltopdf
FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y wkhtmltopdf ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=alpine-stage /app /app

ENV TZ=Asia/Tashkent
RUN ln -snf /usr/share/zoneinfo/Asia/Tashkent /etc/localtime && echo "Asia/Tashkent" > /etc/timezone

RUN chmod +x /app/chatbot

EXPOSE 8080

CMD ["/app/chatbot"]

