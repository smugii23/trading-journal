package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"trading-journal/config"
	"trading-journal/internal/handlers"
)

var db *sql.DB

func init() {
	config.LoadEnv()
	dbURL := config.GetEnv("DATABASE_URL", "")
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
	r.Route("/api", func(r chi.Router) {
		r.Get("/trades", handlers.ListTradesHandler)
		r.Post("/trades", handlers.AddTradeHandler)
		r.Get("/trades/{id}", handlers.GetTradeHandler)
		r.Put("/trades/{id}", handlers.UpdateTradeHandler)
		r.Delete("/trades/{id}", handlers.DeleteTradeHandler)
	})

	// create a handler to serve files from /web/static
	fileServer := http.FileServer(http.Dir("../web/static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	// route for home page/index.html
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../web/templates/index.html")
	})
	// start the server
	log.Fatal(http.ListenAndServe(":8080", r))

}
