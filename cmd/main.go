package main

import (
	"fmt"
	"log"

	"avtest/internal/api"
	"avtest/internal/config"
	"avtest/internal/store/postgres"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("starting service")
	defer logger.Info("service stopped")

	cfg := config.NewConfig()

	r := mux.NewRouter()

	db, err := postgres.NewPostgresDB(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to init db connection: %s", err)
	}

	apiObj := api.NewAPI(logger, r, db)
	apiObj.Run(constructPortString(cfg.APIPort))
}

func constructPortString(port int) string {
	return fmt.Sprintf(":%d", port)
}
