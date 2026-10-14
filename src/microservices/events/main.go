package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

type Payment struct {
	PaymentID  int       `json:"payment_id"`
	UserID     int       `json:"user_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	MethodType string    `json:"method_type"`
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

	mux.HandleFunc("/api/events/movie", handlerMovieEvent)

	mux.HandleFunc("/api/events/user", handlerUserEvent)

	mux.HandleFunc("/api/events/payment", handlerPaymentEvent)

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

func handlerMovieEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - movie event")
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

	log.Printf("UserID - %d : Title - %s : Action - %s MovieID - %d",
		movie.UserID, movie.Title, movie.Action, movie.MovieID)

	// Now `event` is a typed object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func handlerUserEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - user event")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // optional: reject unexpected fields

	if err := dec.Decode(&user); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("user %d (%s) did %s at %s",
		user.UserID, user.Username, user.Action, user.Timestamp)

	// Now `event` is a typed object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handlerPaymentEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - payment event")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payment Payment
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // optional: reject unexpected fields

	if err := dec.Decode(&payment); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("user - %d : PaymentID - %d : Amount is - %.2f : Status - %s at %s with method - %s",
		payment.UserID, payment.PaymentID, payment.Amount, payment.Status, payment.Timestamp, payment.MethodType)

	// Now `event` is a typed object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}
