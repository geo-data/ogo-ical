package main

import (
	"database/sql"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jordic/goics"
	"github.com/lib/pq"
)

// round is taken from <https://github.com/golang/go/issues/4594#issuecomment-135336012>
func round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5, f))
}

// Event represents an ical event.
type Event struct {
	ID             int            `db:"id"`
	Start          time.Time      `db:"start_date"`
	End            time.Time      `db:"end_date"`
	AllDay         bool           `db:"all_day"`
	Title          string         `db:"title"`
	OrganizerName  string         `db:"organizer_name"`
	OrganizerEmail sql.NullString `db:"organizer_email"`
	Attendees      pq.StringArray `db:"attendees"`
	Location       string         `db:"location"`
	Recurrence     sql.NullString `db:"recurrence"`
	Resources      sql.NullString `db:"resources"`
	Comment        sql.NullString `db:"comment"`
	Type           sql.NullString `db:"type"`
	IsPrivate      bool           `db:"is_private"`
}

// DayDuration returns the duration of the event rounded to the number of days.
func (e *Event) DayDuration() int {
	d := e.End.Sub(e.Start)
	return round(d.Hours() / 24)
}

// EventsCollection represents a collection of Event instances.
type EventsCollection []*Event

// EmitICal implements the ICalEmiter interface.
func (ec EventsCollection) EmitICal() goics.Componenter {
	c := goics.NewComponent()
	c.SetType("VCALENDAR")
	c.AddProperty("CALSCAL", "GREGORIAN")
	c.AddProperty("PRODID;X-RICAL-TZSOURCE=TZINFO", "-//geodata.soton.ac.uk")
	c.AddProperty("X-PUBLISHED-TTL", "PT1H") // Format: Duration ([RFC2445] section 4.3.6)
	c.AddProperty("METHOD", "PUBLISH")

	// Generate a component for each event.
	for _, ev := range ec {
		var desc, k, v string
		s := goics.NewComponent()
		s.SetType("VEVENT")

		if ev.AllDay {
			k, v = goics.FormatDateField("DTSTART", ev.Start.In(time.Local))
			s.AddProperty(k, v)
			days := ev.DayDuration()
			if days > 1 {
				end := ev.Start.In(time.Local).Add(time.Duration(days) * (time.Hour * 24))
				k, v = goics.FormatDateField("DTEND", end)
				s.AddProperty(k, v)
			}
		} else {
			k, v = goics.FormatDateTimeField("DTSTART", ev.Start)
			s.AddProperty(k, v)
			k, v = goics.FormatDateTimeField("DTEND", ev.End)
			s.AddProperty(k, v)
		}

		s.AddProperty("UID", strconv.Itoa(ev.ID))

		if ev.OrganizerEmail.Valid && ev.OrganizerEmail.String != "" {
			k = fmt.Sprintf("ORGANIZER;CN=%s", ev.OrganizerName)
			v = fmt.Sprintf("MAILTO:%s", ev.OrganizerEmail.String)
			s.AddProperty(k, v)
		}

		if ev.IsPrivate {
			s.AddProperty("SUMMARY", "Private appointment")

			if ev.Type.Valid {
				cat := strings.Title(ev.Type.String)
				s.AddProperty("CATEGORIES", cat)
			}

		} else {
			s.AddProperty("SUMMARY", ev.Title)

			if len(ev.Location) > 0 {
				s.AddProperty("LOCATION", ev.Location)
			}

			if ev.Type.Valid {
				cat := strings.Title(ev.Type.String)
				s.AddProperty("CATEGORIES", cat)
				desc += fmt.Sprintf("Category: %s\\n", cat)
			}

			if ev.Recurrence.Valid {
				desc += fmt.Sprintf("Recurrence: %s\\n", strings.Title(ev.Recurrence.String))
			}

			/* Attendees shouldn't appear in a PUBLISH calendar.
			for _, a := range ev.Attendees {
				k = fmt.Sprintf("ATTENDEE;CN=%s", a)
				s.AddProperty(k, "MAILTO:geodata@soton.ac.uk")
			}*/
			if len(ev.Attendees) > 0 {
				desc += fmt.Sprintf("Attendees: %s\\n", strings.Join(ev.Attendees, ", "))
			}

			if ev.Resources.Valid {
				s.AddProperty("RESOURCES", ev.Resources.String)
				desc += fmt.Sprintf("Resources: %s\\n", ev.Resources.String)
			}

			if ev.Comment.Valid {
				s.AddProperty("COMMENT", ev.Comment.String)
				desc += fmt.Sprintf("Comment: %s\\n", ev.Comment.String)
			}

			if len(desc) > 0 {
				s.AddProperty("DESCRIPTION", desc)
			}
		}

		c.AddComponent(s)
	}

	return c
}

// Write encodes the events in iCalendar format to w.
func (ec EventsCollection) Write(w io.Writer) {
	goics.NewICalEncode(w).Encode(ec)
}
