package mq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	
	err = ch.ExchangeDeclare("forest_events", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Close() {
	c.ch.Close()
	c.conn.Close()
}

func (c *Client) Publish(ctx context.Context, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.ch.PublishWithContext(ctx,
		"forest_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

func (c *Client) Consume(routingKey string, handler func([]byte) error) error {
	q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = c.ch.QueueBind(q.Name, routingKey, "forest_events", false, nil)
	if err != nil {
		return err
	}

	msgs, err := c.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				log.Printf("Error handling message %s: %v", routingKey, err)
			}
		}
	}()
	return nil
}
