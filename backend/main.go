package main

import (
	"log"
	"net/http"
	"os"

	"polling-app/backend/cache"
	"polling-app/backend/db"
	"polling-app/backend/handlers"
)

func main() {
	database := db.Connect()
	defer database.Close()
	db.RunMigrations(database)

	redisClient := cache.Connect()
	defer redisClient.Close()

	h := &handlers.PollHandler{DB: database, RDB: redisClient}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/polls", h.ListPolls)
	mux.HandleFunc("POST /api/polls", h.CreatePoll)
	mux.HandleFunc("GET /api/polls/{id}", h.GetPoll)
	mux.HandleFunc("POST /api/polls/{id}/vote", h.Vote)
	mux.HandleFunc("GET /api/polls/{id}/has-voted", h.HasVoted)
	mux.HandleFunc("GET /api/polls/{id}/results", h.GetResults)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware(mux)))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
