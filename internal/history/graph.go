package history

import (
	"fmt"
	"math"
	"time"
)

// GraphVisit represents the flattened data structure expected by the D3 visit graph template.
type GraphVisit struct {
	ID             int64   `json:"id"`
	FromVisit      *int64  `json:"from_visit"`
	URL            string  `json:"url"`
	Title          string  `json:"title"`
	CoreTransition int     `json:"core_transition"`
	SecondsOnPage  float64 `json:"seconds_on_page"`
	VisitedAt      string  `json:"visited_at"`
	VisitTS        int64   `json:"visit_ts"` // unix milliseconds, used by the time filter
}

const webkitEpochOffsetSec = 11644473600 // seconds between 1601-01-01 and 1970-01-01

func toWebKit(t time.Time) int64 { return (t.Unix() + webkitEpochOffsetSec) * 1_000_000 }

// GetGraphVisits returns up to `limit` of the most recent visits within [from, to).
// The bool reports whether more visits existed than the limit allowed.
func GetGraphVisits(limit int, from, to *time.Time) ([]GraphVisit, bool, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, false, err
	}
	defer db.Close()

	lo, hi := int64(0), int64(math.MaxInt64)
	if from != nil {
		lo = toWebKit(*from)
	}
	if to != nil {
		hi = toWebKit(*to)
	}

	query := `
		SELECT
		  v.id,
		  v.from_visit,
		  COALESCE(u.url, '')   AS url,
		  COALESCE(u.title, '') AS title,
		  v.transition & 0xFF AS core_transition,
		  v.visit_duration / 1000000.0 AS seconds_on_page,
		  datetime(v.visit_time/1000000 - 11644473600, 'unixepoch', 'localtime') AS visited_at,
		  (v.visit_time / 1000 - 11644473600000) AS visit_ts
		FROM visits v
		JOIN urls u ON u.id = v.url
		WHERE v.visit_time >= ? AND v.visit_time < ?
		ORDER BY v.visit_time DESC
		LIMIT ?;
	`

	// limit <= 0 means "no limit". SQLite treats a negative LIMIT as unlimited.
	// With a limit, ask for one extra row so we can tell "exactly limit" from "more than limit".
	sqlLimit := -1
	if limit > 0 {
		sqlLimit = limit + 1
	}
	rows, err := db.Query(query, lo, hi, sqlLimit)
	if err != nil {
		return nil, false, fmt.Errorf("failed to query graph visits: %w", err)
	}
	defer rows.Close()

	visits := []GraphVisit{}
	for rows.Next() {
		var v GraphVisit
		if err := rows.Scan(
			&v.ID, &v.FromVisit, &v.URL, &v.Title,
			&v.CoreTransition, &v.SecondsOnPage, &v.VisitedAt, &v.VisitTS,
		); err != nil {
			return nil, false, fmt.Errorf("failed to scan graph visit: %w", err)
		}
		visits = append(visits, v)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("error iterating graph visits: %w", err)
	}

	truncated := limit > 0 && len(visits) > limit
	if truncated {
		visits = visits[:limit]
	}
	return visits, truncated, nil
}
