package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
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

// App holds shared dependencies (Kafka producer, config) for all handlers.
type App struct {
	cfg      *Config
	producer sarama.SyncProducer
}

// publish marshals the payload to JSON and sends it to the given topic.
// `key` is used for partition routing; pass "" to let Sarama balance randomly.
func (a *App) publish(topic, key string, payload interface{}) error {
	valueBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(valueBytes),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}

	partition, offset, err := a.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	log.Printf("Message sent to %s [partition=%d offset=%d]", topic, partition, offset)
	return nil
}

func main() {
	fmt.Println("Start Proxy.")
	cfg := loadConfig()

	fmt.Println("__ config start")
	fmt.Println("PORT: ", cfg.Port)
	fmt.Println("KAFKA_BROKERS: ", cfg.KafkaBrokers)
	fmt.Println("__ config end.")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/events/health", handlerHealth)

	//http.HandleFunc("/api/movies", handlerMovies)
	//http.HandleFunc("/api/movies/", handlerMovies)

	// Kafka set
	// 1. Define brokers and configuration
	brokers := []string{cfg.KafkaBrokers}
	config := sarama.NewConfig()

	// Critical: Sarama requires this to be true for sync producers to return success.
	config.Producer.Return.Successes = true

	// Optional: Set Kafka version and required acknowledgments
	config.Version = sarama.V0_10_2_0
	// config.Producer.RequiredAcks = sarama.WaitForAll // Strongest guarantee (acks=all)

	// 2. Create a new synchronous producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	app := &App{cfg: cfg, producer: producer}

	mux.HandleFunc("/api/events/movie", app.handlerMovieEvent)
	mux.HandleFunc("/api/events/user", app.handlerUserEvent)
	mux.HandleFunc("/api/events/payment", app.handlerPaymentEvent)
	// end of kafka set

	log.Printf("Starting events microservice on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("- handled - health check")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func (a *App) handlerMovieEvent(w http.ResponseWriter, r *http.Request) {
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

	// Key by MovieID so all events for the same movie go to the same partition.
	key := strconv.Itoa(movie.MovieID)
	if err := a.publish("movie-events", key, movie); err != nil {
		log.Printf("publish movie failed: %v", err)
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handlerUserEvent(w http.ResponseWriter, r *http.Request) {
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

	key := strconv.Itoa(user.UserID)
	if err := a.publish("user-events", key, user); err != nil {
		log.Printf("publish user failed: %v", err)
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handlerPaymentEvent(w http.ResponseWriter, r *http.Request) {
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

	key := strconv.Itoa(payment.PaymentID)
	if err := a.publish("payment-events", key, payment); err != nil {
		log.Printf("publish payment failed: %v", err)
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
