package main

import (
	"context"
	"encoding/json"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enrichment"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/materialize"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/postgres"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/rabbitmq"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	// Load env config
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initalize postgres
	pgdb, err := postgres.NewPostgresDBFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize Postgres: %v", err)
	}
	defer pgdb.Close()

	// Initalize Redis
	redisConfig := redis.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: "",
		DB:       0,
	}
	redisCache, err := redis.NewRedisCache(redisConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisCache.Close()

	// Initialize RabbitMQ
	rmq, err := rabbitmq.NewRMQ()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rmq.Close()

	playerRepository, err := enrichment.NewPlayerRepository(pgdb)
	if err != nil {
		log.Fatalf("Failed to initialize Player Repository: %v", err)
	}
	defer playerRepository.Close()
	playerService := enrichment.NewPlayerService(playerRepository)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}

	materialized := materialize.NewMaterializer()
	go materialized.StartHTTPServer()

	eventCh := generator.Generate(ctx)

	wg.Add(1)
	go PublishEvents(eventCh, playerService, wg, redisCache, rmq)

	wg.Add(1)
	go SubscribeToEvents(materialized, wg, rmq)

	wg.Wait()

	log.Println("All services finished")
}

func PublishEvents(eventCh <-chan casino.Event, playerService *enrichment.PlayerService, wg *sync.WaitGroup, cache redis.Cache, mq *rabbitmq.RabbitMQ) {
	defer wg.Done()
	for event := range eventCh {
		event = enrichment.GetCommonCurrency(event, cache)
		event.Description = enrichment.GetHumanReadableDescription(event)

		player, err := playerService.GetPlayer(event.PlayerID)
		if err != nil {
			log.Println("Player not found: ", event.PlayerID)
			player = casino.Player{}
		}

		event.Player = player

		ctx := context.Background()
		queueName := os.Getenv("RABBITMQ_EVENT_QUEUE")
		err = mq.Publish(ctx, queueName, event)
		if err != nil {
			log.Println("Failed to publish event", err)
		}
		log.Println("Published event", event)
	}
}

func SubscribeToEvents(materializer *materialize.Materializer, wg *sync.WaitGroup, mq *rabbitmq.RabbitMQ) {
	defer wg.Done()
	queueName := os.Getenv("RABBITMQ_EVENT_QUEUE")
	msg, err := mq.Consume(queueName)
	if err != nil {
		log.Fatalf("Failed to consume events: %v", err)
	}

	for message := range msg {
		var event casino.Event
		err := json.Unmarshal(message.Body, &event)
		if err != nil {
			log.Println("Failed to unmarshal event", err)
			continue
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal materialized event to JSON: %v", err)
		} else {
			log.Printf("Event message consumed: %s", eventJSON)
		}
	}
}
