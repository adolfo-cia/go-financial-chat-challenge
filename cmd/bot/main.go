package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"financial-chat-api/util/config"
	"log"
	"net/http"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	stockCodeTempl = "STOCK_CODE"
	stockApi       = "https://stooq.com/q/l/?s=STOCK_CODE&f=sd2t2ohlcv&h&e=csv"
	receiveQueue   = "financial-sender"
	sendQueue      = "financial-receiver"
)

func main() {
	config := config.Load()
	bot := &bot{
		rabbitUrl:    config.RabbitUrl,
		receiveQueue: receiveQueue,
		sendQueue:    sendQueue,
		send:         make(chan *botMessage),
	}
	go bot.runSender()
	go bot.runReceiver()
	forever := make(chan struct{})
	<-forever
}

type botMessage struct {
	RoomId string `json:"roomId"`
	Msg    string `json:"msg"`
}

type bot struct {
	rabbitUrl    string
	receiveQueue string
	sendQueue    string
	send         chan *botMessage
}

func (b *bot) runReceiver() {
	log.Println("receiver running...")
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

	for d := range msgs {
		log.Printf("Received from queue: %s", d.Body)

		var botMsg botMessage
		err := json.Unmarshal(d.Body, &botMsg)
		if err != nil {
			log.Print("error unmarshaling msg: ", err)
			continue
		}

		b.send <- &botMsg
	}
}

func (b *bot) runSender() {
	log.Println("sender running...")
	conn, err := amqp.Dial(b.rabbitUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	_, err = ch.QueueDeclare(
		b.sendQueue, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range b.send {
		stockValue, err := b.getStockValue(m.Msg)
		if err != nil {
			log.Println("error calling api: ", err)
			continue
		}

		body, err := json.Marshal(&botMessage{RoomId: m.RoomId, Msg: stockValue})
		if err != nil {
			log.Println("error encoding message for rabbitmq", err)
			continue
		}
		err = ch.PublishWithContext(ctx,
			"",          // exchange
			b.sendQueue, // routing key
			false,       // mandatory
			false,       // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf("Sent to queue: %s\n", body)
	}
}

func (b *bot) getStockValue(stock string) (string, error) {
	client := http.DefaultClient //not prod-ready
	resp, err := client.Get(strings.Replace(stockApi, stockCodeTempl, stock, 1))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	csvReader := csv.NewReader(resp.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		return "", nil
	}

	stockValue := records[1][3]

	var sb strings.Builder
	sb.WriteString(strings.ToUpper(stock))
	sb.WriteString(" quote is $")
	sb.WriteString(stockValue)
	sb.WriteString(" per share")
	return sb.String(), nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
