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

/* This is en example of the Config data
   PORT: 8082
   KAFKA_BROKERS: kafka:9092
*/

type Movie struct {
	MovieID int    `json:"movie_id"`
	Title   string `json:"title"`
	Action  string `json:"action"`
	UserID  int    `json:"user_id"`
}

type User struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

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

	mux.HandleFunc("/api/events/movie", handlerMovieCreateEvent)

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

func handlerMovieCreateEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - movie create event")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var movie Movie
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // optional: reject unexpected fields

	if err := dec.Decode(&movie); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("movie %d (%s) did %s at %d",
		movie.UserID, movie.Title, movie.Action, movie.MovieID)

	// Now `event` is a typed object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}
