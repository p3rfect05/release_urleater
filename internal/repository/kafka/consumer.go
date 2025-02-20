package kafkaProducerConsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"time"
	"urleater/dto"
)

type Consumer struct {
	config        KafkaConfig
	consumer      *kafka.Consumer
	workerChannel chan dto.ConsumerData
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

func (c *Consumer) GetWorkerChannel() chan dto.ConsumerData {
	return c.workerChannel
}

func NewConsumer(config KafkaConfig, workerChannel chan dto.ConsumerData) (*Consumer, error) {
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
				err, ok := err.(kafka.Error)

				// Если ошибка является таймаутом, продолжаем цикл
				if !ok {
					log.Println("error while running consumer", err)

					return fmt.Errorf("error while running consumer: %w", err)
				} else {
					if !err.IsTimeout() {
						log.Println("kafka error while running consumer", err)

						return fmt.Errorf("kafka error while running consumer: %w", err)
					} else {
						//log.Println("timeout, continuing...")
						continue
					}
				}
			}

			var data dto.ConsumerData
			if err = json.Unmarshal(msg.Value, &data); err != nil {
				log.Println("error unmarshalling message:", err)

				continue
			}

			log.Println("sending data to worker")

			select {
			case c.workerChannel <- data:
				log.Println("sent data to worker")
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
				log.Println("error committing offsets:", err)
				return fmt.Errorf("failed to commit offsets: %w", err)
			}
			log.Println("committed offset")
		}
	}
}
