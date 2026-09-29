package contextsignal

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"net/url"
	"time"

	"learning-orchestrator/backend/internal/platform"
)

type Server struct {
	DB         *sql.DB
	Experience platform.Client
}
type Event struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Source     string    `json:"source"`
	TrackID    string    `json:"trackId,omitempty"`
	Category   string    `json:"category"`
	ObservedAt time.Time `json:"observedAt"`
	ValidUntil time.Time `json:"validUntil"`
	Canceled   bool      `json:"canceled"`
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health/", platform.Health(s.DB))
	mux.HandleFunc("GET /internal/regions", s.regions)
	mux.HandleFunc("POST /internal/regions", s.saveRegion)
	mux.HandleFunc("DELETE /internal/regions", s.deleteRegions)
	mux.HandleFunc("POST /internal/signals/location", s.location)
	mux.HandleFunc("POST /internal/calendar/ios/windows", func(w http.ResponseWriter, r *http.Request) { s.calendarWindows(w, r, iosCalendar) })
	mux.HandleFunc("DELETE /internal/calendar/ios/windows", func(w http.ResponseWriter, r *http.Request) { s.deleteCalendarWindows(w, r, iosCalendar) })
	mux.HandleFunc("POST /internal/calendar/google/windows", func(w http.ResponseWriter, r *http.Request) { s.calendarWindows(w, r, googleCalendar) })
	mux.HandleFunc("DELETE /internal/calendar/google/windows", func(w http.ResponseWriter, r *http.Request) { s.deleteCalendarWindows(w, r, googleCalendar) })
	mux.HandleFunc("GET /internal/events/{id}", s.event)
	return platform.Internal(mux)
}

func (s Server) regions(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), "SELECT id,version,category,latitude,longitude,radius_meters,track_id::text FROM context.regions WHERE user_id=$1 AND enabled ORDER BY id LIMIT 10", platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load regions")
		return
	}
	defer rows.Close()
	type region struct {
		ID           string  `json:"id"`
		Version      int     `json:"version"`
		Category     string  `json:"category"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		RadiusMeters int     `json:"radiusMeters"`
		TrackID      string  `json:"trackId"`
	}
	items := []region{}
	for rows.Next() {
		var x region
		if rows.Scan(&x.ID, &x.Version, &x.Category, &x.Latitude, &x.Longitude, &x.RadiusMeters, &x.TrackID) != nil {
			platform.Error(w, 500, "database_error", "Could not load regions")
			return
		}
		items = append(items, x)
	}
	if rows.Err() != nil {
		platform.Error(w, 500, "database_error", "Could not load regions")
		return
	}
	platform.JSON(w, 200, map[string]any{"version": 1, "items": items})
}

func (s Server) saveRegion(w http.ResponseWriter, r *http.Request) {
	var in struct {
		InstallationID string  `json:"installationId"`
		TrackID        string  `json:"trackId"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
	}
	if platform.Decode(r, &in) != nil || len(in.InstallationID) == 0 || len(in.InstallationID) > 128 || platform.CheckUUID(in.TrackID) != nil || !validCoordinates(in.Latitude, in.Longitude) {
		platform.Error(w, 422, "validation_failed", "Invalid place coordinates")
		return
	}
	var eligible struct {
		Allowed bool `json:"allowed"`
	}
	if err := s.Experience.Do(r.Context(), platform.User(r), "GET", "/internal/location-eligible?installationId="+url.QueryEscape(in.InstallationID), nil, &eligible); err != nil {
		platform.Error(w, 503, "consent_unavailable", "Cannot verify location permission")
		return
	}
	if !eligible.Allowed {
		platform.Error(w, 403, "location_not_allowed", "Location permission is not active")
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save place")
		return
	}
	defer tx.Rollback()
	allowed, err := locationAllowedTx(r.Context(), tx, platform.User(r), in.InstallationID)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not verify location permission")
		return
	}
	if !allowed {
		platform.Error(w, 403, "location_not_allowed", "Location permission is not active")
		return
	}
	var track string
	err = tx.QueryRowContext(r.Context(), "SELECT id FROM learning.tracks WHERE id=$1 AND user_id=$2 AND status='active' FOR SHARE", in.TrackID, platform.User(r)).Scan(&track)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 422, "unknown_track", "Choose an active learning path")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not verify learning path")
		return
	}
	var id string
	var version int
	err = tx.QueryRowContext(r.Context(), `INSERT INTO context.regions(id,area_id,category,version,latitude,longitude,radius_meters,enabled,user_id,track_id)
 VALUES(gen_random_uuid()::text,'personal','practice_place',1,$2,$3,150,true,$1,$4)
 ON CONFLICT(user_id) WHERE user_id IS NOT NULL DO UPDATE SET latitude=excluded.latitude,longitude=excluded.longitude,
 track_id=excluded.track_id,version=context.regions.version+1,last_exit_at=NULL,enabled=true RETURNING id,version`, platform.User(r), in.Latitude, in.Longitude, track).Scan(&id, &version)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save place")
		return
	}
	_, err = tx.ExecContext(r.Context(), `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source='ios_geofence'
 AND source_resource_id=$2 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, platform.User(r), id)
	if err != nil || tx.Commit() != nil {
		platform.Error(w, 500, "database_error", "Could not save place")
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id, "version": version, "trackId": track, "latitude": in.Latitude, "longitude": in.Longitude, "radiusMeters": 150})
}

func validCoordinates(latitude, longitude float64) bool {
	return latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180 && !math.IsNaN(latitude) && !math.IsNaN(longitude)
}

func (s Server) deleteRegions(w http.ResponseWriter, r *http.Request) {
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove place")
		return
	}
	defer tx.Rollback()
	// Serialize with enrollment and signal ingestion before revoking access.
	var user string
	err = tx.QueryRowContext(r.Context(), "SELECT user_id FROM experience.profiles WHERE user_id=$1 FOR UPDATE", platform.User(r)).Scan(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 500, "database_error", "Could not remove place")
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 SELECT $1,'location_context',latest.policy_version,false,gen_random_uuid()::text FROM (
 SELECT policy_version,granted FROM experience.consents WHERE user_id=$1 AND purpose='location_context'
 ORDER BY decision_seq DESC LIMIT 1) latest WHERE latest.granted`, platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove place")
		return
	}
	if _, err = tx.ExecContext(r.Context(), "UPDATE context.regions SET enabled=false WHERE user_id=$1 AND enabled", platform.User(r)); err != nil {
		platform.Error(w, 500, "database_error", "Could not remove place")
		return
	}
	_, err = tx.ExecContext(r.Context(), `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source='ios_geofence'
 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, platform.User(r))
	if err != nil || tx.Commit() != nil {
		platform.Error(w, 500, "database_error", "Could not remove place")
		return
	}
	platform.JSON(w, 200, map[string]bool{"removed": true})
}

func (s Server) location(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SchemaVersion  int       `json:"schemaVersion"`
		InstallationID string    `json:"installationId"`
		ClientEventID  string    `json:"clientEventId"`
		RegionID       string    `json:"regionId"`
		RegionVersion  int       `json:"regionVersion"`
		Transition     string    `json:"transition"`
		ObservedAt     time.Time `json:"observedAt"`
	}
	if platform.Decode(r, &in) != nil || in.SchemaVersion != 1 || len(in.InstallationID) == 0 || len(in.ClientEventID) == 0 || len(in.ClientEventID) > 128 || len(in.RegionID) == 0 || in.RegionVersion < 1 || (in.Transition != "enter" && in.Transition != "exit") || in.ObservedAt.IsZero() {
		platform.Error(w, 422, "validation_failed", "Invalid location signal")
		return
	}
	now := time.Now()
	if in.ObservedAt.After(now.Add(5*time.Minute)) || in.ObservedAt.Before(now.Add(-30*time.Minute)) {
		platform.Error(w, 422, "stale_signal", "Location signal is stale")
		return
	}
	var eligible struct {
		Allowed bool `json:"allowed"`
	}
	err := s.Experience.Do(r.Context(), platform.User(r), "GET", "/internal/location-eligible?installationId="+url.QueryEscape(in.InstallationID), nil, &eligible)
	if err != nil {
		platform.Error(w, 503, "consent_unavailable", "Cannot verify location permission")
		return
	}
	if !eligible.Allowed {
		platform.Error(w, 403, "location_not_allowed", "Location permission is not active")
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save signal")
		return
	}
	defer tx.Rollback()
	allowed, err := locationAllowedTx(r.Context(), tx, platform.User(r), in.InstallationID)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not verify location permission")
		return
	}
	if !allowed {
		platform.Error(w, 403, "location_not_allowed", "Location permission is not active")
		return
	}
	var category, trackID string
	var lastExit sql.NullTime
	err = tx.QueryRowContext(r.Context(), "SELECT category,last_exit_at,track_id::text FROM context.regions WHERE id=$1 AND version=$2 AND enabled AND user_id=$3 FOR UPDATE", in.RegionID, in.RegionVersion, platform.User(r)).Scan(&category, &lastExit, &trackID)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 422, "unknown_region", "Region is not approved")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not validate region")
		return
	}
	if in.Transition == "exit" {
		if _, err = tx.ExecContext(r.Context(), `UPDATE context.regions SET last_exit_at=CASE
 WHEN last_exit_at IS NULL OR last_exit_at<$2 THEN $2 ELSE last_exit_at END WHERE id=$1`, in.RegionID, in.ObservedAt); err != nil {
			platform.Error(w, 500, "database_error", "Could not cancel context")
			return
		}
		_, err = tx.ExecContext(r.Context(), `WITH canceled AS (
		  UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source='ios_geofence' AND source_resource_id=$2
		  AND observed_at<=$3 AND canceled_at IS NULL RETURNING id
		 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, platform.User(r), in.RegionID, in.ObservedAt)
		if err != nil || tx.Commit() != nil {
			platform.Error(w, 500, "database_error", "Could not cancel context")
			return
		}
		platform.JSON(w, 202, map[string]any{"eventId": nil, "duplicate": false})
		return
	}
	if lastExit.Valid && !in.ObservedAt.After(lastExit.Time) {
		var eventID, storedRegion, storedCategory string
		var storedAt time.Time
		err = tx.QueryRowContext(r.Context(), "SELECT id,source_resource_id,category,observed_at FROM context.events WHERE user_id=$1 AND source='ios_geofence' AND source_event_id=$2", platform.User(r), in.ClientEventID).Scan(&eventID, &storedRegion, &storedCategory, &storedAt)
		if err == nil {
			if storedRegion != in.RegionID || storedCategory != category || !storedAt.Equal(in.ObservedAt) {
				platform.Error(w, 409, "idempotency_conflict", "Event ID was used for a different signal")
				return
			}
			platform.JSON(w, 202, map[string]any{"eventId": eventID, "duplicate": true})
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			platform.Error(w, 500, "database_error", "Could not check signal")
			return
		}
		platform.JSON(w, 202, map[string]any{"eventId": nil, "duplicate": false, "ignored": true})
		return
	}
	var eventID string
	err = tx.QueryRowContext(r.Context(), `INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,track_id,observed_at,valid_until)
 VALUES($1,'ios_geofence',$2,$3,$4,$5,$6,$7) ON CONFLICT(user_id,source,source_event_id) DO NOTHING RETURNING id`, platform.User(r), in.ClientEventID, in.RegionID, category, trackID, in.ObservedAt, in.ObservedAt.Add(30*time.Minute)).Scan(&eventID)
	duplicate := errors.Is(err, sql.ErrNoRows)
	if duplicate {
		var storedRegion, storedCategory string
		var storedAt time.Time
		err = tx.QueryRowContext(r.Context(), "SELECT id,source_resource_id,category,observed_at FROM context.events WHERE user_id=$1 AND source='ios_geofence' AND source_event_id=$2", platform.User(r), in.ClientEventID).Scan(&eventID, &storedRegion, &storedCategory, &storedAt)
		if err == nil && (storedRegion != in.RegionID || storedCategory != category || !storedAt.Equal(in.ObservedAt)) {
			platform.Error(w, 409, "idempotency_conflict", "Event ID was used for a different signal")
			return
		}
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save signal")
		return
	}
	if !duplicate {
		if _, err = tx.ExecContext(r.Context(), "INSERT INTO context.outbox(event_id,kind) VALUES($1,'observe')", eventID); err != nil {
			platform.Error(w, 500, "database_error", "Could not queue signal")
			return
		}
	}
	if err = tx.Commit(); err != nil {
		platform.Error(w, 500, "database_error", "Could not save signal")
		return
	}
	platform.JSON(w, 202, map[string]any{"eventId": eventID, "duplicate": duplicate})
}

// The profile lock serializes this final check with consent revocation.
func locationAllowedTx(ctx context.Context, tx *sql.Tx, user, installation string) (bool, error) {
	var owner string
	err := tx.QueryRowContext(ctx, "SELECT user_id FROM experience.profiles WHERE user_id=$1 FOR UPDATE", user).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM experience.devices d WHERE d.user_id=$1 AND d.installation_id=$2
 AND d.location_permission='enabled') AND COALESCE((SELECT c.granted FROM experience.consents c
 WHERE c.user_id=$1 AND c.purpose='location_context' ORDER BY c.decision_seq DESC LIMIT 1),false)`, user, installation).Scan(&allowed)
	return allowed, err
}

func (s Server) event(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid event")
		return
	}
	var e Event
	var trackID sql.NullString
	err := s.DB.QueryRowContext(r.Context(), "SELECT id,user_id,source,track_id::text,category,observed_at,valid_until,canceled_at IS NOT NULL FROM context.events WHERE id=$1 AND user_id=$2", id, platform.User(r)).Scan(&e.ID, &e.UserID, &e.Source, &trackID, &e.Category, &e.ObservedAt, &e.ValidUntil, &e.Canceled)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 404, "not_found", "Event not found")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load event")
		return
	}
	if trackID.Valid {
		e.TrackID = trackID.String
	}
	platform.JSON(w, 200, e)
}
