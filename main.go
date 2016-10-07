package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/jordic/goics"
	_ "github.com/lib/pq"
)

var (
	config struct {
		Dsn           string
		ServerAddress string
	}
	Db *sqlx.DB
)

func getEnv(name, def string) (value string) {
	if value = os.Getenv(name); value == "" {
		value = def
	}
	return
}

func init() {
	flag.StringVar(&config.Dsn, "dsn", getEnv("OGO_ICAL_DSN", ""), "postgresql Data Source Name")
	flag.StringVar(&config.ServerAddress, "address", getEnv("OGO_ICAL_ADDRESS", ":8080"), "server address")
}

func main() {
	flag.Parse()
	var err error
	Db, err = sqlx.Connect("postgres", config.Dsn)
	if err != nil {
		panic(fmt.Sprintf("Can't connect to database using %s", config.Dsn))
	}

	m := mux.NewRouter()
	m.Handle("/", http.HandlerFunc(CalendarHandler))
	http.Handle("/", m)

	log.Printf("starting server on %s", config.ServerAddress)
	log.Fatal(http.ListenAndServe(config.ServerAddress, nil))
}

func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Calendar request")
	// Setup headers for the calendar
	w.Header().Set("Content-type", "text/calendar")
	w.Header().Set("charset", "utf-8")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("filename", "calendar.ics")

	q := r.URL.Query()
	// Get the Collection models
	collection := GetEvents(q["user"], q["match"])
	// Encode it.
	goics.NewICalEncode(w).Encode(collection)
}
