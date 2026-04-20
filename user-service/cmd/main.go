package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/cache"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/consumer"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/producer"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/service"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting User Service...")

	db, err := initDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer closeDB(db)

	redisClient := initRedis()
	defer closeRedis(redisClient)

	userRepo := repo.NewUserRepo(db)
	userCache := cache.NewUserCache(redisClient)
	userService := service.NewUserService(userRepo, userCache)

	brokers := []string{"kafka:9092"}

	responseProducer := producer.NewResponseProducer(brokers)
	defer closeProducer(responseProducer)

	regConsumer := consumer.NewRegistrationConsumer(brokers, "user-service-reg-group", userRepo)
	if err := regConsumer.Start(context.Background()); err != nil {
		log.Fatalf("Reg consumer failed: %v", err)
	}

	reqConsumer := consumer.NewRequestConsumer(brokers, "user-service-req-group", userService, responseProducer)
	if err := reqConsumer.Start(context.Background()); err != nil {
		log.Fatalf("Req consumer failed: %v", err)
	}

	log.Println("User Service is ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Warning: failed to close database: %v", err)
	}
}

func closeRedis(client *redis.Client) {
	if err := client.Close(); err != nil {
		log.Printf("Warning: failed to close redis: %v", err)
	}
}

func closeProducer(p *producer.ResponseProducer) { // тип подставь точно, как у тебя в producer
	if err := p.Close(); err != nil {
		log.Printf("Warning: failed to close response producer: %v", err)
	}
}

func initDB() (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("POSTGRES_PASSWORD")

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "userdb"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected successfully",
		"host", host,
		"dbname", dbname,
	)

	return db, nil
}

func initRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Warning: Redis ping failed: %v", err)
	}

	log.Println("Redis connected successfully", "addr", addr)

	return client
}
