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

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

func main() {
	config.LoadEnv()
	dbURL := config.GetEnv("DATABASE_URL", "")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	// create a new router
	r := chi.NewRouter()

	// middleware for logging and recovering from panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	tradeHandlers := handlers.NewTradeHandlers(db)
	// define the routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/trades", tradeHandlers.ListTradesHandler)
		r.Post("/trades", tradeHandlers.AddTradeHandler)
		r.Get("/trades/{id}", tradeHandlers.GetTradeHandler)
		r.Put("/trades/{id}", tradeHandlers.UpdateTradeHandler)
		r.Delete("/trades/{id}", tradeHandlers.DeleteTradeHandler)
	})

	// create a handler to serve files from /web/static
	fileServer := http.FileServer(http.Dir("../web/static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	// route for home page/index.html
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../web/templates/index.html")
	})
	// route for adding a trade
	r.Get("/add_trade.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../web/templates/add_trade.html")
	})
	// start the server
	log.Fatal(http.ListenAndServe(":8080", r))

}
