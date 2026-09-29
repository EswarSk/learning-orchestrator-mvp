package contextsignal

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"learning-orchestrator/backend/internal/platform"
)

type calendarWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type calendarSource struct {
	name, purpose, prefix string
}

var iosCalendar = calendarSource{"ios_calendar", "ios_calendar_context", "ios:"}
var googleCalendar = calendarSource{"google_calendar", "google_calendar_context", "google:"}

func validWindows(windows []calendarWindow, now time.Time) bool {
	if len(windows) > 1 {
		return false
	}
	for _, window := range windows {
		if window.Start.IsZero() || window.End.IsZero() || window.Start.Before(now.Add(time.Minute)) ||
			window.Start.After(now.Add(4*time.Minute)) || window.End.Sub(window.Start) < 20*time.Minute ||
			window.End.Sub(window.Start) > 90*time.Minute {
			return false
		}
	}
	return true
}

func calendarAllowedTx(tx *sql.Tx, r *http.Request, installation string, source calendarSource) (bool, error) {
	var owner string
	err := tx.QueryRowContext(r.Context(), "SELECT user_id FROM experience.profiles WHERE user_id=$1 FOR UPDATE", platform.User(r)).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var allowed bool
	err = tx.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM experience.devices d WHERE d.user_id=$1 AND d.installation_id=$2
 AND CASE $3 WHEN 'ios_calendar' THEN d.calendar_permission WHEN 'google_calendar' THEN d.google_calendar_permission ELSE '' END='enabled')
 AND COALESCE((SELECT c.granted FROM experience.consents c
 WHERE c.user_id=$1 AND c.purpose=$4 ORDER BY c.decision_seq DESC LIMIT 1),false)`, platform.User(r), installation, source.name, source.purpose).Scan(&allowed)
	return allowed, err
}

func (s Server) calendarWindows(w http.ResponseWriter, r *http.Request, source calendarSource) {
	var in struct {
		InstallationID string           `json:"installationId"`
		Windows        []calendarWindow `json:"windows"`
	}
	if platform.Decode(r, &in) != nil || len(in.InstallationID) == 0 || len(in.InstallationID) > 128 || in.Windows == nil || !validWindows(in.Windows, time.Now()) {
		platform.Error(w, 422, "validation_failed", "Invalid availability windows")
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save availability")
		return
	}
	defer tx.Rollback()
	allowed, err := calendarAllowedTx(tx, r, in.InstallationID, source)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not verify calendar access")
		return
	}
	if !allowed {
		platform.Error(w, 403, "calendar_not_allowed", "Calendar access is not active")
		return
	}
	resource := source.prefix + in.InstallationID
	rows, err := tx.QueryContext(r.Context(), `SELECT id,evaluate_at,valid_until FROM context.events
 WHERE user_id=$1 AND source=$2 AND source_resource_id=$3 AND canceled_at IS NULL FOR UPDATE`, platform.User(r), source.name, resource)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save availability")
		return
	}
	type existingWindow struct {
		id         string
		start, end time.Time
	}
	existing := []existingWindow{}
	for rows.Next() {
		var item existingWindow
		if err = rows.Scan(&item.id, &item.start, &item.end); err != nil {
			break
		}
		existing = append(existing, item)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save availability")
		return
	}
	kept := make([]bool, len(in.Windows))
	for _, old := range existing {
		match := -1
		for i, current := range in.Windows {
			if !kept[i] && old.start.Equal(current.Start) && old.end.Equal(current.End) {
				match = i
				break
			}
		}
		if match >= 0 {
			kept[match] = true
			_, err = tx.ExecContext(r.Context(), "UPDATE context.events SET observed_at=now() WHERE id=$1", old.id)
			if err != nil {
				break
			}
			continue
		}
		if _, err = tx.ExecContext(r.Context(), `UPDATE context.events SET canceled_at=now() WHERE id=$1`, old.id); err != nil {
			break
		}
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO context.outbox(event_id,kind) VALUES($1,'cancel') ON CONFLICT DO NOTHING`, old.id); err != nil {
			break
		}
	}
	if err == nil {
		for i, current := range in.Windows {
			if kept[i] {
				continue
			}
			var id string
			err = tx.QueryRowContext(r.Context(), `INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,$2,gen_random_uuid()::text,$3,'calendar_availability',now(),$4,$5) RETURNING id`, platform.User(r), source.name, resource, current.Start, current.End).Scan(&id)
			if err != nil {
				break
			}
			_, err = tx.ExecContext(r.Context(), `INSERT INTO context.outbox(event_id,kind) VALUES($1,'observe')`, id)
			if err != nil {
				break
			}
		}
	}
	if err != nil || tx.Commit() != nil {
		platform.Error(w, 500, "database_error", "Could not save availability")
		return
	}
	platform.JSON(w, 202, map[string]any{"accepted": len(in.Windows)})
}

func (s Server) deleteCalendarWindows(w http.ResponseWriter, r *http.Request, source calendarSource) {
	installation := r.URL.Query().Get("installationId")
	if len(installation) == 0 || len(installation) > 128 {
		platform.Error(w, 422, "validation_failed", "Invalid installation")
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove availability")
		return
	}
	defer tx.Rollback()
	var owner string
	err = tx.QueryRowContext(r.Context(), "SELECT user_id FROM experience.profiles WHERE user_id=$1 FOR UPDATE", platform.User(r)).Scan(&owner)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove availability")
		return
	}
	var bound bool
	err = tx.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM experience.devices WHERE user_id=$1 AND installation_id=$2)", owner, installation).Scan(&bound)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove availability")
		return
	}
	if !bound {
		platform.Error(w, 403, "device_not_bound", "Device does not belong to this account")
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE experience.devices SET calendar_permission=CASE WHEN $3='ios_calendar' THEN 'denied' ELSE calendar_permission END,
 google_calendar_permission=CASE WHEN $3='google_calendar' THEN 'denied' ELSE google_calendar_permission END
 WHERE user_id=$1 AND installation_id=$2`, owner, installation, source.name)
	if err == nil {
		err = cancelCalendarResource(r, tx, owner, source.name, source.prefix+installation)
	}
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 SELECT $1,$2,latest.policy_version,false,gen_random_uuid()::text FROM (
 SELECT policy_version,granted FROM experience.consents WHERE user_id=$1 AND purpose=$2
 ORDER BY decision_seq DESC LIMIT 1) latest WHERE latest.granted AND NOT EXISTS (
 SELECT 1 FROM experience.devices WHERE user_id=$1 AND
 CASE $3 WHEN 'ios_calendar' THEN calendar_permission WHEN 'google_calendar' THEN google_calendar_permission ELSE '' END='enabled')`, owner, source.purpose, source.name)
	}
	if err != nil || tx.Commit() != nil {
		platform.Error(w, 500, "database_error", "Could not remove availability")
		return
	}
	platform.JSON(w, 200, map[string]bool{"removed": true})
}

func cancelCalendarResource(r *http.Request, tx *sql.Tx, user, source, resource string) error {
	_, err := tx.ExecContext(r.Context(), `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source=$2
 AND source_resource_id=$3 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, user, source, resource)
	return err
}
