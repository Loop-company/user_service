package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/cache"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/consumer"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/eventbus"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/service"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	usergrpc "github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/grpc"
	userpb "github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/proto"
	"google.golang.org/grpc"
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

	brokers := getKafkaBrokers()
	kafkaProducer := eventbus.NewKafkaProducer(brokers, slog.Default())

	userService := service.NewUserService(userRepo, userCache, kafkaProducer)

	grpcServer := grpc.NewServer()

	userServer := usergrpc.NewUserServer(userService)

	userpb.RegisterUserServiceServer(grpcServer, userServer)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server started on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	regConsumer := consumer.NewRegistrationConsumer(brokers, "user-service-reg-group", userRepo)
	if err := regConsumer.Start(context.Background()); err != nil {
		log.Fatalf("Reg consumer failed: %v", err)
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

func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	return strings.Split(brokers, ",")
}
