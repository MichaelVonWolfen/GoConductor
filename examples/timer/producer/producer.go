package main

import (
	"GoConductor/rsc/AnsiColors"
	"encoding/json"
	"flag"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math/rand"
	"os"
)

import (
	"context"
	"time"
)

var ProdLog *log.Logger

func failOnError(err error, msg string) {
	if err != nil {
		ProdLog.Panicf("%s: %s", msg, err)
	}
}

type Message struct {
	Body  string
	Sleep int
}

func main() {
	name := flag.String("name", "producer", "Program Name")
	flag.Parse()
	ProdLog = log.New(os.Stdout, fmt.Sprintf("%s%s:%s", AnsiColors.BlueText, *name, AnsiColors.ResetText), log.LstdFlags)

	ProdLog.Println(os.Args)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	callsNumber := 10

	for i := 0; i < callsNumber; i++ {
		message := Message{
			Body:  fmt.Sprintf("Hello number %d!", i),
			Sleep: rand.Intn(10),
		}
		body, err := json.Marshal(message)
		failOnError(err, "Failed to marshal JSON")
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		failOnError(err, "Failed to publish a message")
		ProdLog.Printf(" [x] Sent %s\n to queue: %s\n", body, q.Name)
	}
	//for i := range 11 {
	//	for j := range 10 {
	//		n := 10*i + j
	//		if n > 108 {
	//			break
	//		}
	//		fmt.Printf("\033[%dm %3d\033[m", n, n)
	//	}
	//	fmt.Println()
	//}
}
