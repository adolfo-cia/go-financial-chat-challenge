package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type botMessage struct {
	RoomId string `json:"roomId"`
	Msg    string `json:"msg"`
}

type bot struct {
	amqpCh         *amqp.Channel
	rabbitUrl      string
	sendQueue      string
	receiveQueue   string
	registerRoomId chan uuid.UUID
	sendCh         chan *botMessage
	roomReceiveCh  map[uuid.UUID]chan *botMessage
}

func NewBot(amqpCh *amqp.Channel, rabbitUrl string, sendQueue string, receiveQueue string) *bot {
	return &bot{
		amqpCh:         amqpCh,
		rabbitUrl:      rabbitUrl,
		sendQueue:      sendQueue,
		receiveQueue:   receiveQueue,
		registerRoomId: make(chan uuid.UUID),
		sendCh:         make(chan *botMessage),
		roomReceiveCh:  make(map[uuid.UUID]chan *botMessage),
	}
}

func (b *bot) Run() {
	log.Println("bot running...")
	conn, err := amqp.Dial(b.rabbitUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	_, err = ch.QueueDeclare(
		b.receiveQueue, // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		b.receiveQueue, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	failOnError(err, "Failed to register a consumer")

	for {
		select {
		case d := <-msgs:
			log.Printf("received from queue: %s", d.Body)
			var botMsg botMessage
			err := json.Unmarshal(d.Body, &botMsg)
			if err != nil {
				log.Print("error unmarshaling msg: ", err)
				continue
			}
			roomUuid, err := uuid.Parse(botMsg.RoomId)
			if err != nil {
				log.Print("error creating uuid: ", err)
				continue
			}
			b.roomReceiveCh[roomUuid] <- &botMsg
		case m := <-b.sendCh:
			b.send(m)
		case roomId := <-b.registerRoomId:
			b.roomReceiveCh[roomId] = make(chan *botMessage)
		}
	}
}

func (b *bot) send(message *botMessage) {
	stockCode, err := b.extractStockTicker(message.Msg)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(&botMessage{RoomId: message.RoomId, Msg: stockCode})
	if err != nil {
		log.Println("error encoding message for rabbitmq", err)
		return
	}

	err = b.amqpCh.PublishWithContext(ctx,
		"",          // exchange
		b.sendQueue, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Printf("failed to publish message: %s", err)
		return
	}

	log.Printf("Sent to queue: %s\n", body)
}

func (b *bot) extractStockTicker(msg string) (string, error) {
	const prefix = "/stock="
	if strings.HasPrefix(msg, prefix) {
		return strings.TrimPrefix(msg, prefix), nil
	}
	return "", fmt.Errorf("input string does not start with %s", prefix)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
