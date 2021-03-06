package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	reaper "github.com/ramr/go-reaper"
)

var (
	// config stores the program configuration.
	config struct {
		Dsn           string
		ServerAddress string
	}
	showVersion     bool
	version, commit string
)

// getEnv returns the value of the environment variable name or def if the
// variable is empty.
func getEnv(name, def string) (value string) {
	if value = os.Getenv(name); value == "" {
		value = def
	}
	return
}

// handleSignals handles any termination signals.
func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Exiting on %s signal", sig.String())
	os.Exit(1)
}

// connect attempts to connect to the data store a number of times in case of
// temporary network errors.
func connect(store *Store, maxAttempts int) (err error) {
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		if err = store.Connect(); err == nil {
			break
		}

		// If it's a temporary network error, try again.
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			duration := time.Duration(attempts) * time.Second
			log.Printf("Network error connecting to the database (attempt %d of %d): trying again in %s", attempts, maxAttempts, duration)
			time.Sleep(duration)
			continue
		} else {
			return
		}
	}

	return
}

func init() {
	// Initialise the flag options.
	flag.BoolVar(&showVersion, "version", false, "display version information")
	flag.StringVar(&config.Dsn, "dsn", getEnv("OGO_ICAL_DSN", ""), "postgresql Data Source Name")
	flag.StringVar(&config.ServerAddress, "address", getEnv("OGO_ICAL_ADDRESS", ":8080"), "server address")
}

func main() {
	//  Start background reaping of orphaned child processes.
	go reaper.Reap()

	flag.Parse()

	if showVersion {
		if version != "" && commit != "" {
			fmt.Printf("%s commit=%s\n", version, commit)
		} else {
			fmt.Println("No versioning information is available.")
		}
		os.Exit(0)
	}

	go handleSignals()

	// Connect to the data source.
	store := NewStore(config.Dsn)
	if err := connect(store, 20); err != nil {
		log.Fatal(err)
	}

	// Start the HTTP server.
	handler := CalendarHandler(store)
	log.Printf("Starting server on %s", config.ServerAddress)
	log.Fatal(http.ListenAndServe(config.ServerAddress, handler))
}
