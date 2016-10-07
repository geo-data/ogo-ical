package main

import (
	"log"
	"net/http"

	"github.com/jordic/goics"
)

// CalendarHandler creates a http.Handler for dealing with calendar requests.
func CalendarHandler(store *Store) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Set up iCalendar headers.
		w.Header().Set("Content-type", "text/calendar")
		w.Header().Set("charset", "utf-8")
		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("filename", "calendar.ics")

		q := r.URL.Query()
		// Get the matching events.
		if collection, err := store.Events(q["user"], q["match"]); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
		} else {
			// Encode the collection.
			goics.NewICalEncode(w).Encode(collection)
		}
	}

	return http.HandlerFunc(handler)
}
