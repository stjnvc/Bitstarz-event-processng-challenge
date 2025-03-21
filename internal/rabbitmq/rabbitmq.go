package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewRMQ() (*RabbitMQ, error) {
	rmq := &RabbitMQ{}
	err := rmq.connect()
	if err != nil {
		return nil, err
	}
	return rmq, nil
}

func (r *RabbitMQ) connect() error {
	RmqUrl := os.Getenv("RABBITMQ_URL")
	if RmqUrl == "" {
		log.Fatalf("Failed to read RMQ_URL from environment variable")
	}

	conn, err := amqp091.Dial(RmqUrl)
	if err != nil {
		return err
	}
	r.conn = conn

	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	r.channel = ch

	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, queueName string, message interface{}) error {
	q, err := r.channel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(message)

	if err != nil {
		return err
	}

	msg := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp091.Persistent,
	}

	err = r.channel.PublishWithContext(ctx, "", q.Name, false, false, msg)

	return err
}

func (r *RabbitMQ) Consume(queueName string) (<-chan amqp091.Delivery, error) {
	q, err := r.channel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	msgs, err := r.channel.Consume(q.Name, "", true, false, false, false, nil)

	return msgs, err
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		err := r.channel.Close()
		if err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
}
