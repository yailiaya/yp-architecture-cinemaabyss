package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Config — конфигурация прокси-сервиса, собирается из переменных окружения.
type Config struct {
	Port         string
	KafkaBrokers string
}

/*
   PORT: 8000
   MONOLITH_URL: http://monolith:8080
   #монолит
   MOVIES_SERVICE_URL: http://movies-service:8081 #сервис movies
   EVENTS_SERVICE_URL: http://events-service:8082
   GRADUAL_MIGRATION: "true" # вкл/выкл простого фиче-флага
   MOVIES_MIGRATION_PERCENT: "50" # процент миграции
*/

func loadConfig() *Config {
	fmt.Println("__ reading config:")
	port := os.Getenv("PORT") // "" если не задано
	kafka_brokers := os.Getenv("KAFKA_BROKERS")

	fmt.Println("PORT: ", port)
	fmt.Println("KAFKA_BROKERS: ", kafka_brokers)
	fmt.Println("__ end of config.")

	return &Config{
		Port:         port,
		KafkaBrokers: kafka_brokers,
	}
}

func main() {
	fmt.Println("Start Proxy.")
	cfg := loadConfig()

	fmt.Println("__ config start")
	fmt.Println("PORT: ", cfg.Port)
	fmt.Println("KAFKA_BROKERS: ", cfg.KafkaBrokers)
	fmt.Println("__ config end.")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlerHealth)

	//http.HandleFunc("/api/movies", handlerMovies)
	//http.HandleFunc("/api/movies/", handlerMovies)

	log.Printf("Starting events microservice on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - health check")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}
