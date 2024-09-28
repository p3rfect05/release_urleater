package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/antonlindstrom/pgstore"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urleater/internal/config"
	"urleater/internal/handlers"
	"urleater/internal/repository/postgresDB"
	"urleater/internal/service"
	"urleater/internal/validator"
)

const port = ":8080"

func main() {
	serverCtx, serverCancel := context.WithCancel(context.Background())
	postgresConfig := config.Config{
		DB: config.DBConfig{
			PostgresHost:     "localhost",
			PostgresPort:     "5432",
			PostgresUser:     "postgres",
			PostgresPassword: "postgres",
			PostgresDatabase: "postgres",
			PostgresParams:   "sslmode=disable",
		}}
	postgresPool := providePool(serverCtx, postgresConfig.PostgresURL(), true)

	// storage layer
	postgresStorage := postgresDB.NewStorage(postgresPool)

	// service layer
	srv := service.New(postgresStorage)

	store, err := pgstore.NewPGStore(postgresConfig.PostgresURL(), []byte("secret-key")) // TODO make env for secret key

	sessionStore := handlers.NewPostgresSessionStore(store)

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer store.Close()

	defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// handlers layer
	e := handlers.GetRoutes(&handlers.Handlers{Service: srv, Store: sessionStore})

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
