package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"trading-journal/internal/handlers"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// Get DATABASE_URL from environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
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
	tagHandlers := handlers.NewTagHandlers(db)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// define the routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/trades", tradeHandlers.ListTradesHandler)
		r.Post("/trades", tradeHandlers.AddTradeHandler)
		r.Get("/trades/{id}", tradeHandlers.GetTradeHandler)
		r.Put("/trades/{id}", tradeHandlers.UpdateTradeHandler)
		r.Delete("/trades/{id}", tradeHandlers.DeleteTradeHandler)

		r.Get("/tags", tagHandlers.ListTagsHandler)
		r.Post("/tags", tagHandlers.CreateTagHandler)
		r.Get("/tags/{id}", tagHandlers.GetTagHandler)
		r.Put("/tags/{id}", tagHandlers.UpdateTagHandler)
		r.Delete("/tags/{id}", tagHandlers.DeleteTagHandler)

		r.Get("/trades/{trade_id}/tags", tagHandlers.GetTradeTagsHandler)
		r.Post("/trades/{trade_id}/tags/{tag_id}", tagHandlers.AddTagToTradeHandler)
		r.Delete("/trades/{trade_id}/tags/{tag_id}", tagHandlers.RemoveTagFromTradeHandler)
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
