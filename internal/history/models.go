package history

import (
	"time"
)

// URL represents a row from the "urls" table.
type URL struct {
	Id            int64     `db:"id"`
	Url           string    `db:"url"`
	Title         string    `db:"title"`
	VisitCount    int       `db:"visit_count"`
	TypedCount    int       `db:"typed_count"`
	LastVisitTime time.Time `db:"last_visit_time"`
	Hidden        bool      `db:"hidden"`
}

// Visit represents a complete 1:1 mapping of Chrome's "visits" table schema.
type Visit struct {
	Id                           int64     `db:"id"`
	UrlId                        int64     `db:"url_id"` // Foreign key referencing urls.id
	VisitTime                    time.Time `db:"visit_time"`
	FromVisit                    *int64    `db:"from_visit"`
	ExternalReferrerURL          *string   `db:"external_referrer_url"`
	Transition                   int       `db:"transition"`
	SegmentID                    *int64    `db:"segment_id"`
	VisitDuration                int64     `db:"visit_duration"` // Microseconds spent on page
	IncrementedOmniboxTypedScore bool      `db:"incremented_omnibox_typed_score"`
	OpenerVisit                  *int64    `db:"opener_visit"`
	OriginatorCacheGuid          *string   `db:"originator_cache_guid"`
	OriginatorVisitId            *int64    `db:"originator_visit_id"`
	OriginatorFromVisit          *int64    `db:"originator_from_visit"`
	OriginatorOpenerVisit        *int64    `db:"originator_opener_visit"`
	IsKnownToSync                bool      `db:"is_known_to_sync"`
	ConsiderForNtpMostVisited    bool      `db:"consider_for_ntp_most_visited"`
	VisitedLinkId                int64     `db:"visited_link_id"`
	AppId                        *string   `db:"app_id"`
}

// VisitNode combines URL and Visit data together, forming a complete record
// ready to be serialized into nodes/edges for your visit graph visualization.
// type VisitNode struct {
// 	VisitID       int64     `json:"visit_id"`
// 	URL           string    `json:"url"`
// 	Title         string    `json:"title"`
// 	VisitTime     time.Time `json:"visit_time"`
// 	FromVisit     *int64    `json:"from_visit,omitempty"`   // Points to parent Visit ID
// 	OpenerVisit   *int64    `json:"opener_visit,omitempty"` // Points to opener Visit ID
// 	VisitDuration int64     `json:"visit_duration_ms"`      // Converted to milliseconds for convenience
// }

// VisitEdge represents a directional link between two page visits (Parent -> Child)
// perfect for rendering a graph network (e.g., using D3.js, vis.js, or Cytoscape.js).
// type VisitEdge struct {
// 	Source int64 `json:"source"` // Parent visit ID (from_visit)
// 	Target int64 `json:"target"` // Child visit ID (current visit ID)
// }
