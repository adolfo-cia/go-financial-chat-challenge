package main

import (
	"context"
	db "financial-chat-api/db/sqlc"
	"financial-chat-api/internal/chat"
	"financial-chat-api/internal/user"
	"financial-chat-api/util/auth"
	"financial-chat-api/util/config"
	"financial-chat-api/util/password"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tomiok/webh"
)

const (
	sendQueue    = "financial-sender"
	receiveQueue = "financial-receiver"
)

func main() {

	config := config.Load()

	dbpool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatalln("unable to connect to db", err)
	}
	defer dbpool.Close()

	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalln("cannot ping to db", err)
	}

	db := db.New(dbpool)

	tokenMaker, err := auth.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatalln("error creating tokenMaker", err)
	}

	userRepo := user.NewRepository(db)
	userServ := user.NewService(userRepo, password.Hash, password.Check, tokenMaker)
	userHandler := user.NewHandler(userServ)

	conn, err := amqp.Dial(config.RabbitUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	_, err = ch.QueueDeclare(
		sendQueue, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	bot := chat.NewBot(ch, config.RabbitUrl, sendQueue, receiveQueue)
	go bot.Run()
	hub := chat.NewHub(bot)
	go hub.Run()
	chatHandler := chat.NewHandler(hub)

	server := webh.NewServer(
		config.ServerPort,
		webh.WithHeartbeat("/ping"),
		webh.WithRequestLogger("financial-chat",
			httplog.Options{
				JSON:    false,
				Concise: true,
			}))

	server.Post("/users", webh.Unwrap(userHandler.CreateUser))
	server.Post("/login", webh.Unwrap(userHandler.Login))

	server.Route("/", func(r chi.Router) {
		r.Use(auth.Middleware(tokenMaker))
		r.Post("/rooms", webh.Unwrap(chatHandler.HandleCreateRoom))
		r.Get("/ws", webh.Unwrap(chatHandler.HandleJoinRoom))
	})

	server.Start()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
