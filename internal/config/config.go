package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type DBConfig struct {
	PostgresHost     string `envconfig:"postgres_host" required:"true"`
	PostgresPort     string `envconfig:"postgres_port" required:"true"`
	PostgresUser     string `envconfig:"postgres_user" required:"true"`
	PostgresPassword string `envconfig:"postgres_password" required:"true"`
	PostgresDatabase string `envconfig:"postgres_database" required:"true"`
	PostgresParams   string `envconfig:"postgres_params" required:"false"`
}

type RedisConfig struct {
	Host string `envconfig:"redis_host" required:"true"`
	Port string `envconfig:"redis_port" required:"true"`
}

type ElasticConfig struct {
	Host string `envconfig:"elastic_host" required:"true"`
}

type Config struct {
	DB                     DBConfig
	Redis                  RedisConfig
	Elastic                ElasticConfig
	Kafka                  KafkaConfig
	ConsumingWorkersNumber int `envconfig:"consuming_workers_number" required:"true" default:"100"`
}

type KafkaConfigConsumer struct {
	GroupId           string `envconfig:"kafka_group_id" required:"true"`
	Topic             string `envconfig:"kafka_topics" required:"true"`
	NumberOfConsumers int    `envconfig:"number_of_consumers" required:"true"`
}

type KafkaConfigProducer struct {
	Topic string `envconfig:"kafka_topics" required:"true"`
}

type KafkaConfig struct {
	Address            string `envconfig:"kafka_address" required:"true"`
	Consumer           KafkaConfigConsumer
	Producer           KafkaConfigProducer
	NumberOfPartitions int `envconfig:"kafka_number_of_partitions" required:"false" default:"1"`
	ReplicationFactor  int `envconfig:"kafka_replication_factor" required:"false" default:"1"`
}

func ProvideConfig() *Config {
	cfg, err := FromEnv()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func FromEnv() (*Config, error) {
	cfg := new(Config)

	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("error while parse env config | %w", err)
	}

	return cfg, nil
}

func (c *Config) PostgresURL() string {
	pgURL := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		c.DB.PostgresUser,
		c.DB.PostgresPassword,
		c.DB.PostgresHost,
		c.DB.PostgresPort,
		c.DB.PostgresDatabase,
	)

	if c.DB.PostgresParams != "" {
		pgURL = fmt.Sprintf("%v?%v", pgURL, c.DB.PostgresParams)
	}

	return pgURL
}
