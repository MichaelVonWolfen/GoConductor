package main

import (
	"GoConductor/rsc/AnsiColors"
	"encoding/json"
	"flag"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"time"
)

var TimerLog *log.Logger

func failOnError(err error, msg string) {
	if err != nil {
		TimerLog.Panicf("%s: %s", msg, err)
	}
}

type Message struct {
	Body  string
	Sleep int
}

func main() {
	name := flag.String("name", "Timer", "Program Name")
	flag.Parse()
	TimerLog = log.New(os.Stdout, fmt.Sprintf("%s%s:%s", AnsiColors.MagentaText, *name, AnsiColors.ResetText), log.LstdFlags)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"hello", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var body Message
			err := json.Unmarshal(d.Body, &body)
			failOnError(err, "Failed to parse a message")
			failOnError(err, "Failed to parse a string to number")
			TimerLog.Printf("Received the msg: %s and will sleep for %d", body.Body, body.Sleep)
			time.Sleep(time.Duration(body.Sleep) * time.Second)
		}
	}()

	TimerLog.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
