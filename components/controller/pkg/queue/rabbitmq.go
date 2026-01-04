package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/davidliyutong/idekube-controller/internal/config"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQClient(cfg config.RabbitMQConfig) (*RabbitMQClient, error) {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
	}, nil
}

func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *RabbitMQClient) Channel() *amqp.Channel {
	return c.channel
}

// TODO: Add message queue operation methods
// Example:
// func (c *RabbitMQClient) Publish(ctx context.Context, queue string, message []byte) error
// func (c *RabbitMQClient) Consume(ctx context.Context, queue string) (<-chan amqp.Delivery, error)
