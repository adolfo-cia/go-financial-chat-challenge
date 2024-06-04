# migration related lines are commented, as docker compose uses a dependent service to execute migration
# for building this Dockerfile, uncomment those lines before executing 'docker build'

# build stage
FROM golang:1.22.3-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/chat/main.go
# RUN apk add curl
# RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

# run stage
FROM alpine:3.19 AS runner
WORKDIR /app
# COPY --from=builder /app/migrate .
# COPY db/migration ./migration
# COPY start.sh .
# RUN chmod +x start.sh
COPY .env .
COPY --from=builder /app/main .

EXPOSE 8080
# CMD [ "/app/main" ]
# ENTRYPOINT [ "/app/start.sh" ]
ENTRYPOINT [ "/app/main" ]
