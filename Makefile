network:
	docker network create financial-chat-network

rbmq-pull:
	docker pull rabbitmq:3.13.3-alpine

rbmq-run:
	docker run --name financial-bot-queue --network=financial-chat-network -p 5672:5672 --hostname financial-bot-queue rabbitmq:3.13.3-alpine

postgres-pull:
	docker pull postgres:14.11-alpine3.19

postgres-run:
	docker run --name financial-chat-db --network=financial-chat-network -p 5432:5432 -e POSTGRES_PASSWORD=root -e POSTGRES_USER=root -e POSTGRES_DB=financial-chat -d postgres:14.11-alpine3.19

createdb:
	docker exec -it financial-chat-db createdb --username=root --owner=root financial-chat

dropdb:
	docker exec -it financial-chat-db dropdb financial-chat

sqlc:
	sqlc generate

migrate-new:
	migrate create -ext sql -dir db/migration -seq init_schema
# add new migration: migrate create -ext sql -dir db/migration -seq add_users

migrateup:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose down 1

server:
	go run main.go

financial-chat-build:
	docker build -f chat.Dockerfile -t financial-chat-api:latest .

financial-chat-run:
	docker run --name financial-chat-api --network=financial-chat-network -p 8080:8080 -e DB_SOURCE="postgresql://root:root@financial-chat-db:5432/financial-chat?sslmode=disable" financial-chat-api:latest

financial-bot-build:
	docker build -f bot.Dockerfile -t financial-bot:latest .

financial-bot-run:
	docker run --name financial-bot --network=financial-chat-network -p 8080:8080 -e RABBIT_URL="amqp://guest:guest@financial-bot-queue:5672/" financial-bot:latest


.PHONY: network rbmq-pull rbmq-run postgres-pull postgres-run postgres-start postgres-stop createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc server financial-chat-build financial-chat-run financial-bot-build financial-bot-run