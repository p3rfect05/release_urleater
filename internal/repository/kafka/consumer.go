package kafkaProducerConsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"time"
)

type Consumer struct {
	config        KafkaConfig
	consumer      *kafka.Consumer
	workerChannel chan ConsumerData
}

type KafkaConfig struct {
	KafkaConfig  *kafka.ConfigMap
	KafkaTopics  []string
	KafkaGroupId string
	KafkaServer  string
}

func (c *Consumer) GetConfig() KafkaConfig {
	return c.config
}

func (c *Consumer) GetWorkerChannel() chan ConsumerData {
	return c.workerChannel
}

func NewConsumer(config KafkaConfig, workerChannel chan ConsumerData) (*Consumer, error) {
	newConsumer, err := kafka.NewConsumer(config.KafkaConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating kafka consumer: %w", err)
	}

	err = newConsumer.SubscribeTopics(config.KafkaTopics, nil)

	if err != nil {
		return nil, fmt.Errorf("error subscribing to topics: %w", err)
	}

	return &Consumer{
		config:        config,
		consumer:      newConsumer,
		workerChannel: workerChannel,
	}, nil
}

func (c *Consumer) StartConsuming(ctx context.Context) error {
	defer c.consumer.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.consumer.ReadMessage(500 * time.Millisecond)

			if err != nil {
				var kafkaError *kafka.Error

				ok := errors.As(err, &kafkaError)

				if !ok || ok && !kafkaError.IsTimeout() { // либо не ошибка кафки, либо ошибка кафки но не IsTimeout
					return fmt.Errorf("error reading message: %w", err)
				}

				continue
			}

			var data ConsumerData

			err = json.Unmarshal(msg.Value, &data)

			if err != nil {
				continue
			}

			select {
			case c.workerChannel <- data:
			case <-ctx.Done():
				return nil
			}

			_, err = c.consumer.CommitOffsets([]kafka.TopicPartition{
				{
					Topic:     msg.TopicPartition.Topic,
					Partition: msg.TopicPartition.Partition,
					Offset:    msg.TopicPartition.Offset + 1,
				},
			})

			if err != nil {
				return fmt.Errorf("failed to commit offsets: %w", err)
			}
		}
	}
}
