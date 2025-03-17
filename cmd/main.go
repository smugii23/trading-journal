package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"trading-journal/internal/handlers"
)

var db *sql.DB

func init() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	var err error
	connStr := dbURL
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// check the connection is alive
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// create a new router
	r := chi.NewRouter()

	// middleware for logging and recovering from panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// define the routes
	r.Get("/trades", handlers.ListTradesHandler)
	r.Post("/trades", handlers.AddTradeHandler)
	r.Get("/trades/{id}", handlers.GetTradeHandler)
	r.Put("/trades/{id}", handlers.UpdateTradeHandler)
	r.Delete("/trades/{id}", handlers.DeleteTradeHandler)
	// start the server
	log.Fatal(http.ListenAndServe(":8080", r))

}
