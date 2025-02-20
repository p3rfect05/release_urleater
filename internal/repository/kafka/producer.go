package kafkaProducerConsumer

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"urleater/dto"
)

type Producer struct {
	config   KafkaConfig
	producer *kafka.Producer
}

func NewProducer(config KafkaConfig) (*Producer, error) {
	newProducer, err := kafka.NewProducer(config.KafkaConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating kafka producer: %w", err)
	}

	return &Producer{
		config:   config,
		producer: newProducer,
	}, nil
}

func (p *Producer) GetConfig() KafkaConfig {
	return p.config
}

func (p *Producer) PublishMsg(msgType string, data any, topic string) error {
	var err error

	defer func() {
		if err != nil {
			p.Close()
		}
	}()

	byteData, err := json.Marshal(dto.ConsumerData{
		TypeOfMessage: msgType,
		Data:          data,
	})

	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: byteData,
	}, nil)

	if err != nil {
		return fmt.Errorf("error publishing data: %w", err)
	}

	return nil
}

func (p *Producer) Close() {
	if p.producer != nil {
		p.producer.Close()
	}
}
