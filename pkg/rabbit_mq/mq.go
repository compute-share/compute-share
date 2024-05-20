package mq

import (
	"context"
	// "fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	mqSingleton *RabbitMQ
)

// Singleton instance
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

type MessageHandler func(msg []byte)

// Helper
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// GetJobQueueSingleton returns a singleton instance of RabbitMQ
func GetJobQueueSingleton() *RabbitMQ {
	if (mqSingleton != nil) {
		return mqSingleton
	}
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"job-queue", 
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	mqSingleton = &RabbitMQ{
		conn:    conn,
		channel: ch,
		queue:   q,
	}

	return mqSingleton
}

func (r *RabbitMQ) PublishMessage(body string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.channel.PublishWithContext(ctx,
		"",         // exchange
		r.queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}

func (r *RabbitMQ) ConsumeMessages(handler MessageHandler) {
    msgs, err := r.channel.Consume(
        r.queue.Name, // queue
        "",           // consumer
        true,         // auto-ack
        false,        // exclusive
        false,        // no-local
        false,        // no-wait
        nil,          // args
    )
    if err != nil {
        log.Fatalf("Failed to register a consumer: %s", err)
    }

    go func() {
        for msg := range msgs {
            handler(msg.Body)
        }
    }()

    log.Printf(" [*] Waiting for jobs. To exit press CTRL+C")

	forever := make(chan bool)
    <-forever
}

func (r *RabbitMQ) Close() {
    r.channel.Close()
    r.conn.Close()
}
