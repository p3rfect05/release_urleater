package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/antonlindstrom/pgstore"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urleater/dto"
	"urleater/internal/config"
	"urleater/internal/handlers"
	"urleater/internal/repository/elastic_searcher"
	kafkaProducerConsumer "urleater/internal/repository/kafka"
	"urleater/internal/repository/postgresDB"
	"urleater/internal/repository/redisDB"
	"urleater/internal/service"
	"urleater/internal/validator"
)

const port = ":8080"

func main() {
	serverCtx, serverCancel := context.WithCancel(context.Background())

	cfg := config.ProvideConfig()

	postgresPool := providePool(serverCtx, cfg.PostgresURL(), true)

	// storage layer
	postgresStorage := postgresDB.NewStorage(postgresPool)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Host + ":" + cfg.Redis.Port,
	})

	redisStorage := redisDB.NewStorage(redisClient)

	elasticClient, err := elastic.NewClient(elastic.SetURL(cfg.Elastic.Host))

	if err != nil {
		log.Fatal(err)
	}

	elasticSearcher := elastic_searcher.NewSearcher(elasticClient, cfg.Elastic.Host)

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Address,
		"group.id":          cfg.Kafka.Consumer.GroupId,
		"auto.offset.reset": "earliest",
	}

	kafkaConfig := kafkaProducerConsumer.KafkaConfig{
		KafkaConfig:  configMap,
		KafkaTopics:  []string{cfg.Kafka.Consumer.Topic},
		KafkaGroupId: cfg.Kafka.Consumer.GroupId,
		KafkaServer:  cfg.Kafka.Address,
	}

	admin, err := kafka.NewAdminClient(configMap)

	if err != nil {
		log.Fatalf("Failed to create Admin client: %v", err)
	}

	defer admin.Close()

	// Задаем спецификацию топика, который нужно создать.
	topicSpec := kafka.TopicSpecification{
		Topic:             cfg.Kafka.Consumer.Topic,
		NumPartitions:     cfg.Kafka.NumberOfPartitions,
		ReplicationFactor: cfg.Kafka.ReplicationFactor,
	}

	// Создаем топик. Если топик уже существует, ошибка будет проигнорирована.
	results, err := admin.CreateTopics(serverCtx, []kafka.TopicSpecification{topicSpec})

	if err != nil {
		log.Fatalf("Error creating topics: %v", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Fatalf("Failed to create topic %s: %v", result.Topic, result.Error)
		} else {
			log.Printf("Topic %s created or already exists", result.Topic)
		}
	}

	workerChannel := make(chan dto.ConsumerData, 100000)

	consumers := make([]service.Consumer, 0, cfg.Kafka.Consumer.NumberOfConsumers)

	for i := 0; i < cfg.Kafka.Consumer.NumberOfConsumers; i++ {
		consumer, err := kafkaProducerConsumer.NewConsumer(kafkaConfig, workerChannel)

		if err != nil {
			log.Fatalf("Failed to create consumer: %v", err)
		}

		consumers = append(consumers, consumer)
	}

	// Создаем продюсера.
	producer, err := kafkaProducerConsumer.NewProducer(kafkaConfig)

	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}

	// service layer
	srv := service.New(postgresStorage, redisStorage, producer, consumers, elasticSearcher, cfg.Kafka.Producer.Topic)

	store, err := pgstore.NewPGStore(cfg.PostgresURL(), []byte("secret-key")) // TODO make env for secret key

	sessionStore := handlers.NewPostgresSessionStore(store)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer store.Close()

	defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// handlers layer
	e := handlers.GetRoutes(&handlers.Handlers{Service: srv, Store: sessionStore})

	err = srv.CreateSubscriptions(serverCtx)

	if err != nil {
		log.Println(err.Error())
	}

	srv.StartConsumers(serverCtx)

	srv.StartConsumingWorkers(serverCtx, cfg.ConsumingWorkersNumber, workerChannel)

	httpValidator, err := validator.NewValidator()

	if err != nil {
		panic(err)
	}

	e.Validator = httpValidator

	go func() {
		if err := e.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	fmt.Println("Shutting down server...")

	serverCancel()
}

func providePool(ctx context.Context, url string, lazy bool) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(url)

	if err != nil {
		log.Fatal("Unable to parse DB config because " + err.Error())
	}

	poolConfig.LazyConnect = lazy

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("Unable to establish connection to " + url + " because " + err.Error())
	}

	return pool
}
