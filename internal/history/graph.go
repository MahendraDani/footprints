package history

import (
	"fmt"
)

// GraphVisit represents the flattened data structure expected by the D3 visit graph template.
type GraphVisit struct {
	ID             int64   `json:"id"`
	FromVisit      *int64  `json:"from_visit"`
	URL            string  `json:"url"` // must be exported, or encoding/json ignores it
	Title          string  `json:"title"`
	CoreTransition int     `json:"core_transition"`
	SecondsOnPage  float64 `json:"seconds_on_page"`
	VisitedAt      string  `json:"visited_at"`
}

// GetGraphVisits queries the database and transforms raw history rows into graph-ready nodes.
func GetGraphVisits(limit int) ([]GraphVisit, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// COALESCE guards against NULLs: urls.url and urls.title are nullable in Chrome's schema,
	// and scanning NULL into a plain string makes rows.Scan fail.
	query := `
		SELECT
		  v.id,
		  v.from_visit,
		  COALESCE(u.url, '')   AS url,
		  COALESCE(u.title, '') AS title,
		  v.transition & 0xFF AS core_transition,
		  v.visit_duration / 1000000.0 AS seconds_on_page,
		  datetime(v.visit_time/1000000 - 11644473600, 'unixepoch', 'localtime') AS visited_at
		FROM visits v
		JOIN urls u ON u.id = v.url
		ORDER BY v.visit_time DESC
		LIMIT ?;
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query graph visits: %w", err)
	}
	defer rows.Close()

	visits := make([]GraphVisit, 0, limit) // non-nil, so an empty result marshals to [] instead of null
	for rows.Next() {
		var v GraphVisit
		if err := rows.Scan(
			&v.ID,
			&v.FromVisit,
			&v.URL,
			&v.Title,
			&v.CoreTransition,
			&v.SecondsOnPage,
			&v.VisitedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan graph visit: %w", err)
		}
		visits = append(visits, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating graph visits: %w", err)
	}

	return visits, nil
}
