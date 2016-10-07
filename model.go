package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jordic/goics"
	"github.com/lib/pq"
)

// Event represents an ical event.
type Event struct {
	ID         int            `db:"id"`
	Start      time.Time      `db:"start_date"`
	End        time.Time      `db:"end_date"`
	Title      string         `db:"title"`
	Attendees  pq.StringArray `db:"attendees"`
	Location   string         `db:"location"`
	Recurrence sql.NullString `db:"recurrence"`
	Resources  sql.NullString `db:"resources"`
	Comment    sql.NullString `db:"comment"`
	Type       sql.NullString `db:"type"`
}

// A collection of rous
type EventsCollection []*Event

// We implement ICalEmiter interface that will return a goics.Componenter.
func (ec EventsCollection) EmitICal() goics.Componenter {
	c := goics.NewComponent()
	c.SetType("VCALENDAR")
	c.AddProperty("CALSCAL", "GREGORIAN")
	c.AddProperty("PRODID;X-RICAL-TZSOURCE=TZINFO", "-//geodata.soton.ac.uk")

	for _, ev := range ec {
		var desc string
		s := goics.NewComponent()
		s.SetType("VEVENT")
		s.AddProperty("SUMMARY", ev.Title)

		k, v := goics.FormatDateTimeField("DTSTART", ev.Start)
		s.AddProperty(k, v)
		k, v = goics.FormatDateTimeField("DTEND", ev.End)
		s.AddProperty(k, v)

		s.AddProperty("UID", strconv.Itoa(ev.ID))

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

		for _, a := range ev.Attendees {
			k = fmt.Sprintf("ATTENDEE;CN=%s", a)
			s.AddProperty(k, "MAILTO:geodata@soton.ac.uk")
		}
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

		c.AddComponent(s)
	}

	return c
}

// Get data from database populating EventsCollection
func GetEvents(users, keywords []string) EventsCollection {
	var (
		userpos, kwpos []string
		userq, kwq     string
		params         []interface{}
		pos            int
	)

	for _, user := range users {
		pos += 1
		params = append(params, user)
		userpos = append(userpos, fmt.Sprintf("c.login = $%d", pos))
	}

	if len(userpos) > 0 {
		userq = fmt.Sprintf("AND (%s)", strings.Join(userpos, " OR "))
	}

	for _, kw := range keywords {
		pos += 1
		params = append(params, "%"+kw+"%")
		kwpos = append(kwpos, fmt.Sprintf("e.title ILIKE $%d", pos))
	}

	if len(kwpos) > 0 {
		kwq = fmt.Sprintf("AND (%s)", strings.Join(kwpos, " OR "))
	}

	q := fmt.Sprintf(`WITH RECURSIVE logins(n) AS (
  SELECT c.company_id AS n
    FROM company c,
         date_company_assignment dc,
         events
    WHERE c.company_id = dc.company_id
      AND c.is_team = 1
      AND dc.date_id = events.date_id
  UNION ALL
  SELECT c.company_id
    FROM company c,
         logins,
         company_assignment ca
    WHERE c.company_id = ca.sub_company_id
      AND logins.n = ca.company_id
      %s
), events AS (
  SELECT d.start_date,
         d.end_date,
         d.date_id,
         d.title,
         di.comment,
         d.location,
         d.resource_names,
         d.type,
         d.apt_type,
         (SELECT array_agg(CASE WHEN c.firstname IS NULL AND c.name IS NULL
                                THEN c.description
                                ELSE (c.firstname || ' ' || c.name) END)
            FROM company c,
                 date_company_assignment dc
           WHERE d.date_id = dc.date_id
             AND c.company_id = dc.company_id) AS attendees
    FROM date_x d, date_info di
    WHERE d.date_id = di.date_id
      AND d.start_date::date >= now()::date
)
SELECT DISTINCT
       e.date_id AS id,
       e.title,
       e.attendees,
       e.start_date,
       e.end_date,
       e.location,
       e.comment,
       e.resource_names AS resources,
       e.type AS recurrence,
       e.apt_type AS type
  FROM logins l,
       company c,
       date_company_assignment dca,
       events e
  WHERE e.date_id = dca.date_id
    AND dca.company_id = c.company_id
    AND l.n = c.company_id
    %s
  ORDER BY e.start_date DESC;`, userq, kwq)

	events := EventsCollection{}
	err := Db.Select(&events, q, params...)
	_ = Db.Unsafe()
	if err != nil {
		log.Println(err)
	}
	return events
}
