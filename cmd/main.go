package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	dbURL := os.Getenv("DATABASE_URL")
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
	// start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
