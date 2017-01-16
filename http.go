package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// CalendarHandler creates a http.Handler for dealing with calendar requests.
func CalendarHandler(store *Store) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// Limit requests to those including a user and or match.  This prevents
		// downloads of an entire calendar database.
		if len(q["user"]) == 0 && len(q["match"]) == 0 {
			http.Error(w, "Forbidden.", 403)
			return
		}

		// Get the matching events.
		if collection, err := store.Events(q["user"], q["match"]); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
		} else {
			// Set up iCalendar headers.
			w.Header().Set("Content-type", "text/calendar")
			w.Header().Set("charset", "utf-8")
			w.Header().Set("Content-Disposition", "inline")
			w.Header().Set("filename", "calendar.ics")

			// Encode the collection.
			collection.Write(w)
		}

		return
	}

	return handlers.LoggingHandler(os.Stdout, http.HandlerFunc(handler))
}
