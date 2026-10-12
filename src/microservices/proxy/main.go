package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

/*
   PORT: 8000
   MONOLITH_URL: http://monolith:8080
   #монолит
   MOVIES_SERVICE_URL: http://movies-service:8081 #сервис movies
   EVENTS_SERVICE_URL: http://events-service:8082
   GRADUAL_MIGRATION: "true" # вкл/выкл простого фиче-флага
   MOVIES_MIGRATION_PERCENT: "50" # процент миграции
*/

func main() {
	fmt.Println("Start Proxy.")

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

	http.HandleFunc("/api/proxy/health", handlerHealth)
	//http.HandleFunc("/api/users", handleUsers)
	//http.HandleFunc("/api/movies/health", handlerHealth)

	log.Printf("Starting movies microservice on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}
