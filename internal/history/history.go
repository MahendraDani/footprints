package history

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/MahendraDani/footprints/internal/workspace"
	_ "modernc.org/sqlite"
)

// OpenDB establishes a safe, read-only connection to the cached History database.
// This function is reusable across any query you write in the future.
func OpenDB() (*sql.DB, error) {
	historyPath, err := workspace.HistoryPath()
	if err != nil {
		return nil, err
	}

	// Ensure the cached database file actually exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("history file not found at %s. Please run 'fp init' first", historyPath)
	}

	// Open using modernc.org/sqlite driver in read-only mode (mode=ro prevents locks & corruption)
	dsn := fmt.Sprintf("file:%s?mode=ro", historyPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %w", err)
	}

	// Verify the connection is actually alive
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to history database: %w", err)
	}

	return db, nil
}

// GetRecentVisits is a reusable query function that pulls recent visits using the full model.
func GetRecentVisits(limit int) ([]Visit, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT 
			v.id, v.url_id, v.visit_time, v.from_visit, v.external_referrer_url, 
			v.transition, v.segment_id, v.visit_duration, v.incremented_omnibox_typed_score, 
			v.opener_visit, v.originator_cache_guid, v.originator_visit_id, 
			v.originator_from_visit, v.originator_opener_visit, v.is_known_to_sync, 
			v.consider_for_ntp_most_visited, v.visited_link_id, v.app_id
		FROM visits v
		ORDER BY v.visit_time DESC
		LIMIT ?;
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query visits: %w", err)
	}
	defer rows.Close()

	var visits []Visit
	for rows.Next() {
		var v Visit
		var chromeTime int64

		err := rows.Scan(
			&v.Id, &v.UrlId, &chromeTime, &v.FromVisit, &v.ExternalReferrerURL,
			&v.Transition, &v.SegmentID, &v.VisitDuration, &v.IncrementedOmniboxTypedScore,
			&v.OpenerVisit, &v.OriginatorCacheGuid, &v.OriginatorVisitId,
			&v.OriginatorFromVisit, &v.OriginatorOpenerVisit, &v.IsKnownToSync,
			&v.ConsiderForNtpMostVisited, &v.VisitedLinkId, &v.AppId,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan visit row: %w", err)
		}

		// Convert Chromium microsecond timestamp to Go time.Time
		v.VisitTime = chromeTimeToTime(chromeTime)
		visits = append(visits, v)
	}

	return visits, nil
}

// Helper: Converts Chromium epoch (microseconds since Jan 1, 1601) to standard time.Time
func chromeTimeToTime(chromeTime int64) time.Time {
	if chromeTime == 0 {
		return time.Time{}
	}
	unixMicro := chromeTime - 11644473600000000
	return time.Unix(0, unixMicro*int64(time.Microsecond))
}
