package opportunity

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"learning-orchestrator/backend/internal/platform"
)

type Server struct{ DB *sql.DB }
type Offer struct {
	ID             string    `json:"id"`
	TrackID        string    `json:"trackId"`
	NodeID         string    `json:"nodeId"`
	ContextEventID string    `json:"contextEventId"`
	ContextLabel   string    `json:"contextLabel"`
	ReasonCode     string    `json:"reasonCode"`
	ValidUntil     time.Time `json:"validUntil"`
	Status         string    `json:"status"`
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health/", platform.Health(s.DB))
	mux.HandleFunc("GET /internal/offers", s.list)
	mux.HandleFunc("GET /internal/offers/{id}", s.get)
	mux.HandleFunc("POST /internal/offers/{id}/actions", s.action)
	mux.HandleFunc("POST /internal/offers/by-context/{id}/cancel", s.cancelByContext)
	mux.HandleFunc("POST /internal/offers/{id}/complete", s.complete)
	return platform.Internal(mux)
}

func (s Server) complete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid opportunity")
		return
	}
	var status string
	err := s.DB.QueryRowContext(r.Context(), `UPDATE opportunity.offers SET status='completed',updated_at=now()
 WHERE id=$1 AND user_id=$2 AND status IN ('ready','completed') RETURNING status`, id, platform.User(r)).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 409, "state_conflict", "Opportunity cannot be completed")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not complete opportunity")
		return
	}
	platform.JSON(w, 200, map[string]string{"status": status})
}

func (s Server) cancelByContext(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid context event")
		return
	}
	_, err := s.DB.ExecContext(r.Context(), "UPDATE opportunity.offers SET status='canceled',updated_at=now() WHERE context_event_id=$1 AND user_id=$2 AND status IN ('ready','offered','scheduled')", id, platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not cancel opportunity")
		return
	}
	platform.JSON(w, 200, map[string]string{"status": "canceled"})
}

func (s Server) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT o.id,o.track_id,o.node_id,o.context_label,o.reason_code,o.valid_until,
 CASE WHEN o.status='scheduled' AND o.snooze_until<=now() THEN 'ready' ELSE o.status END AS status
 FROM opportunity.offers o JOIN context.events e ON e.id=o.context_event_id AND e.user_id=o.user_id
 JOIN experience.profiles p ON p.user_id=o.user_id
 WHERE o.user_id=$1 AND o.valid_until>now() AND e.canceled_at IS NULL AND e.valid_until>now()
 AND context.calendar_sources_agree(e.id,o.user_id)
 AND NOT p.proactive_paused AND COALESCE((SELECT c.granted FROM experience.consents c WHERE c.user_id=o.user_id
 AND c.purpose=CASE e.source WHEN 'ios_geofence' THEN 'location_context' WHEN 'calendar' THEN 'calendar_context' WHEN 'ios_calendar' THEN 'ios_calendar_context' WHEN 'google_calendar' THEN 'google_calendar_context' ELSE '' END
 ORDER BY c.decision_seq DESC LIMIT 1),false)
 AND (o.status IN ('ready','offered') OR (o.status='scheduled' AND o.snooze_until<=now()))
 ORDER BY o.created_at DESC LIMIT 50`, platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load opportunities")
		return
	}
	defer rows.Close()
	items := []Offer{}
	for rows.Next() {
		var x Offer
		if rows.Scan(&x.ID, &x.TrackID, &x.NodeID, &x.ContextLabel, &x.ReasonCode, &x.ValidUntil, &x.Status) != nil {
			platform.Error(w, 500, "database_error", "Could not load opportunities")
			return
		}
		items = append(items, x)
	}
	if rows.Err() != nil {
		platform.Error(w, 500, "database_error", "Could not load opportunities")
		return
	}
	platform.JSON(w, 200, map[string]any{"items": items, "nextCursor": nil})
}

func (s Server) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid opportunity")
		return
	}
	var x Offer
	err := s.DB.QueryRowContext(r.Context(), `SELECT id,track_id,node_id,context_event_id,context_label,reason_code,valid_until,
 CASE WHEN status='scheduled' AND snooze_until<=now() THEN 'ready' ELSE status END FROM opportunity.offers WHERE id=$1 AND user_id=$2`, id, platform.User(r)).Scan(&x.ID, &x.TrackID, &x.NodeID, &x.ContextEventID, &x.ContextLabel, &x.ReasonCode, &x.ValidUntil, &x.Status)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 404, "not_found", "Opportunity not found")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load opportunity")
		return
	}
	platform.JSON(w, 200, x)
}

func (s Server) action(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid opportunity")
		return
	}
	var in struct {
		Action      string     `json:"action"`
		SnoozeUntil *time.Time `json:"snoozeUntil"`
	}
	if platform.Decode(r, &in) != nil || (in.Action != "dismiss" && in.Action != "snooze") || (in.Action == "snooze" && in.SnoozeUntil == nil) {
		platform.Error(w, 422, "validation_failed", "Invalid action")
		return
	}
	if in.SnoozeUntil != nil && !in.SnoozeUntil.After(time.Now()) {
		platform.Error(w, 422, "validation_failed", "Snooze time must be in the future")
		return
	}
	state := "dismissed"
	if in.Action == "snooze" {
		state = "scheduled"
	}
	var x Offer
	err := s.DB.QueryRowContext(r.Context(), `UPDATE opportunity.offers SET status=$3,snooze_until=$4,updated_at=now() WHERE id=$1 AND user_id=$2 AND valid_until>now()
 AND ($4::timestamptz IS NULL OR $4<valid_until)
 AND status IN ('ready','offered','scheduled') RETURNING id,track_id,node_id,context_label,reason_code,valid_until,status`, id, platform.User(r), state, in.SnoozeUntil).Scan(&x.ID, &x.TrackID, &x.NodeID, &x.ContextLabel, &x.ReasonCode, &x.ValidUntil, &x.Status)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 409, "state_conflict", "Opportunity cannot be changed")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not update opportunity")
		return
	}
	platform.JSON(w, 200, x)
}
