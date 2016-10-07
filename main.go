package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	reaper "github.com/ramr/go-reaper"
)

// config stores the program configuration.
var config struct {
	Dsn           string
	ServerAddress string
}

// getEnv returns the value of the environment variable name or def if the
// variable is empty.
func getEnv(name, def string) (value string) {
	if value = os.Getenv(name); value == "" {
		value = def
	}
	return
}

func init() {
	// Initialise the flag options.
	flag.StringVar(&config.Dsn, "dsn", getEnv("OGO_ICAL_DSN", ""), "postgresql Data Source Name")
	flag.StringVar(&config.ServerAddress, "address", getEnv("OGO_ICAL_ADDRESS", ":8080"), "server address")
}

func main() {
	//  Start background reaping of orphaned child processes.
	go reaper.Reap()

	flag.Parse()

	// Connect to the data source.
	store := NewStore(config.Dsn)
	if err := store.Connect(); err != nil {
		log.Fatal(err)
	}

	// Start the HTTP server.
	handler := CalendarHandler(store)
	log.Printf("Starting server on %s", config.ServerAddress)
	log.Fatal(http.ListenAndServe(config.ServerAddress, handler))
}
