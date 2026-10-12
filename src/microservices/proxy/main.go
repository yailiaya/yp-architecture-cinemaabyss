package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Config — конфигурация прокси-сервиса, собирается из переменных окружения.
type Config struct {
	Port                   string
	MonolithURL            string
	MoviesServiceURL       string
	EventsServiceURL       string
	GradualMigration       bool
	MoviesMigrationPercent int
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
	monolith_url := os.Getenv("MONOLITH_URL")
	movies_url := os.Getenv("MOVIES_SERVICE_URL")
	events_url := os.Getenv("EVENTS_SERVICE_URL")
	gradual_switch := os.Getenv("GRADUAL_MIGRATION")

	proxy_percent := os.Getenv("MOVIES_MIGRATION_PERCENT")

	fmt.Println("PORT: ", port)
	fmt.Println("MONOLITH_URL: ", monolith_url)
	fmt.Println("MOVIES_SERVICE_URL: ", movies_url)
	fmt.Println("EVENTS_SERVICE_URL: ", events_url)
	fmt.Println("GRADUAL_MIGRATION: ", gradual_switch)
	fmt.Println("MOVIES_MIGRATION_PERCENT: ", proxy_percent)
	fmt.Println("__ end of config.")

	gradualMigrationSwitch, _ := strconv.ParseBool(gradual_switch)
	proxyPercent, _ := strconv.Atoi(proxy_percent)

	return &Config{
		Port:                   port,
		MonolithURL:            monolith_url,
		MoviesServiceURL:       movies_url,
		EventsServiceURL:       events_url,
		GradualMigration:       gradualMigrationSwitch,
		MoviesMigrationPercent: proxyPercent,
	}
}

func main() {
	fmt.Println("Start Proxy.")
	cfg := loadConfig()

	fmt.Println("__ config start")
	fmt.Println("PORT: ", cfg.Port)
	fmt.Println("MONOLITH_URL: ", cfg.MonolithURL)
	fmt.Println("MOVIES_SERVICE_URL: ", cfg.MoviesServiceURL)
	fmt.Println("EVENTS_SERVICE_URL: ", cfg.EventsServiceURL)
	fmt.Println("GRADUAL_MIGRATION: ", cfg.GradualMigration)
	fmt.Println("MOVIES_MIGRATION_PERCENT: ", cfg.MoviesMigrationPercent)
	fmt.Println("__ config end.")

	http.HandleFunc("/api/proxy/health", handlerHealth)

	//http.HandleFunc("/api/movies", handlerMovies)
	//http.HandleFunc("/api/movies/", handlerMovies)

	log.Printf("Starting movies microservice on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

/*
func makeHandler(сап *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// db доступна здесь
		rows, _ := db.Query("SELECT ...")
		_ = rows
	}
}

func handlerMovies(w http.ResponseWriter, r *http.Request) {
	r.Host := movies_url
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}
*/
