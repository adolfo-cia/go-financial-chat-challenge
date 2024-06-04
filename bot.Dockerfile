# build stage
FROM golang:1.22.3-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/bot/main.go

# run stage
FROM alpine:3.19 AS runner
WORKDIR /app
COPY .env .
COPY --from=builder /app/main .

EXPOSE 8081
ENTRYPOINT [ "/app/main" ]
