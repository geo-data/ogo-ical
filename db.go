package main

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Store represents a data source.
type Store struct {
	dsn string
	db  *sqlx.DB
}

// NewStore creates a new Store.
func NewStore(dsn string) *Store {
	return &Store{dsn, nil}
}

// Connect connects to the database.
func (s *Store) Connect() (err error) {
	s.db, err = sqlx.Connect("postgres", s.dsn)
	return
}

// Events returns an EventsCollection.  Events are filtered to match users and keywords.
func (s *Store) Events(users, keywords []string) (events EventsCollection, err error) {
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
  SELECT ((((d.start_date AT TIME ZONE 'UTC')::timestamp without time zone)::timestamp with time zone) AT TIME ZONE 'UTC') AS start_date,
         ((((d.end_date AT TIME ZONE 'UTC')::timestamp without time zone)::timestamp with time zone) AT TIME ZONE 'UTC') AS end_date,
         ((d.start_date AT TIME ZONE 'UTC')::time = '00:00:00'::time AND (d.end_date AT TIME ZONE 'UTC')::time = '23:59:00'::time) AS all_day,
         d.date_id,
         d.title,
         di.comment,
         d.location,
         d.resource_names,
         d.type,
         d.apt_type,
         d.access_team_id IS NULL AS is_private,
         CASE WHEN c.firstname IS NULL AND c.name IS NULL
              THEN c.description
              ELSE (c.firstname || ' ' || c.name) END AS organizer_name,
         cv.value_string AS organizer_email,
         (SELECT array_agg(CASE WHEN c.firstname IS NULL AND c.name IS NULL
                                THEN c.description
                                ELSE (c.firstname || ' ' || c.name) END)
            FROM company c,
                 date_company_assignment dc
           WHERE d.date_id = dc.date_id
             AND c.company_id = dc.company_id) AS attendees
    FROM date_x d, date_info di, company c, company_value cv
    WHERE d.date_id = di.date_id
      AND (d.start_date AT TIME ZONE 'UTC')::date >= (CURRENT_DATE - interval '2 mon')::date
      AND c.company_id = d.owner_id
      AND cv.company_id = c.company_id
      AND cv.attribute = 'email1'
)
SELECT DISTINCT
       e.date_id AS id,
       e.title,
       e.attendees,
       e.organizer_name,
       e.organizer_email,
       e.start_date,
       e.end_date,
       e.all_day,
       e.location,
       e.comment,
       e.is_private,
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

	err = s.db.Select(&events, q, params...)
	return
}
