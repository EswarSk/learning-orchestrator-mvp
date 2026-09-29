package experience

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	_ "time/tzdata"

	"learning-orchestrator/backend/internal/platform"
)

type Server struct {
	DB                             *sql.DB
	Auth                           Verifier
	Learning, Context, Opportunity platform.Client
}
type profile struct {
	UserID          string `json:"userId"`
	Timezone        string `json:"timezone"`
	ProactivePaused bool   `json:"proactivePaused"`
	QuietStart      string `json:"quietStart"`
	QuietEnd        string `json:"quietEnd"`
}

func lockProfile(ctx context.Context, tx *sql.Tx, user string) error {
	if _, err := tx.ExecContext(ctx, "INSERT INTO experience.profiles(user_id) VALUES($1) ON CONFLICT DO NOTHING", user); err != nil {
		return err
	}
	var found string
	return tx.QueryRowContext(ctx, "SELECT user_id FROM experience.profiles WHERE user_id=$1 FOR UPDATE", user).Scan(&found)
}

func (s Server) Handler() http.Handler {
	root := http.NewServeMux()
	root.Handle("GET /health/", platform.Health(s.DB))
	internal := http.NewServeMux()
	internal.HandleFunc("GET /internal/location-eligible", s.locationEligible)
	internal.HandleFunc("GET /internal/opportunity-eligible", s.opportunityEligible)
	root.Handle("/internal/", platform.Internal(internal))
	public := http.NewServeMux()
	public.HandleFunc("GET /v1/bootstrap", s.bootstrap)
	public.HandleFunc("GET /v1/subjects", s.subjects)
	public.HandleFunc("GET /v1/tracks", s.tracks)
	public.HandleFunc("POST /v1/tracks", s.createTrack)
	public.HandleFunc("GET /v1/curriculum", s.curriculum)
	public.HandleFunc("GET /v1/progress", s.progress)
	public.HandleFunc("PATCH /v1/me", s.patchMe)
	public.HandleFunc("GET /v1/me/export", s.exportData)
	public.HandleFunc("POST /v1/consents", s.consent)
	public.HandleFunc("PUT /v1/devices/{installationId}", s.device)
	public.HandleFunc("DELETE /v1/devices/{installationId}", s.deleteDevice)
	public.HandleFunc("GET /v1/notifications", s.notificationStatus)
	public.HandleFunc("GET /v1/location/regions", s.regions)
	public.HandleFunc("POST /v1/location/regions", s.saveRegion)
	public.HandleFunc("DELETE /v1/location/regions", s.deleteRegions)
	public.HandleFunc("POST /v1/signals/location", s.location)
	public.HandleFunc("POST /v1/calendar/ios/windows", func(w http.ResponseWriter, r *http.Request) { s.calendarWindows(w, r, "ios") })
	public.HandleFunc("DELETE /v1/calendar/ios/windows", func(w http.ResponseWriter, r *http.Request) { s.deleteCalendarWindows(w, r, "ios") })
	public.HandleFunc("POST /v1/calendar/google/windows", func(w http.ResponseWriter, r *http.Request) { s.calendarWindows(w, r, "google") })
	public.HandleFunc("DELETE /v1/calendar/google/windows", func(w http.ResponseWriter, r *http.Request) { s.deleteCalendarWindows(w, r, "google") })
	public.HandleFunc("GET /v1/opportunities", s.offers)
	public.HandleFunc("POST /v1/opportunities/{id}/actions", s.offerAction)
	public.HandleFunc("POST /v1/opportunities/{id}/applications", s.application)
	public.HandleFunc("POST /v1/sessions", s.startSession)
	public.HandleFunc("POST /v1/sessions/{id}/turns", s.turn)
	public.HandleFunc("POST /v1/sessions/{id}/speech", s.voice)
	public.HandleFunc("POST /v1/sessions/{id}/transcribe", s.voice)
	public.HandleFunc("POST /v1/sessions/{id}/realtime", s.realtime)
	public.HandleFunc("POST /v1/sessions/{id}/finish", s.finish)
	public.HandleFunc("GET /v1/billing/status", s.billing)
	root.Handle("/v1/", s.Auth.Middleware(public))
	return root
}

func (s Server) loadProfile(r *http.Request) (profile, error) {
	var p profile
	err := s.DB.QueryRowContext(r.Context(), "SELECT timezone,proactive_paused,to_char(quiet_start,'HH24:MI'),to_char(quiet_end,'HH24:MI') FROM experience.profiles WHERE user_id=$1", platform.User(r)).Scan(&p.Timezone, &p.ProactivePaused, &p.QuietStart, &p.QuietEnd)
	p.UserID = platform.User(r)
	return p, err
}

func (s Server) bootstrap(w http.ResponseWriter, r *http.Request) {
	_, err := s.DB.ExecContext(r.Context(), "INSERT INTO experience.profiles(user_id) VALUES($1) ON CONFLICT DO NOTHING", platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not create profile")
		return
	}
	p, err := s.loadProfile(r)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load profile")
		return
	}
	var learning map[string]any
	if err = s.Learning.Do(r.Context(), platform.User(r), "GET", "/internal/summary"+trackQuery(r), nil, &learning); err != nil {
		remoteError(w, err)
		return
	}
	learning["learner"] = p
	learning["consentSummary"] = []any{}
	learning["connectionSummary"] = []any{}
	platform.JSON(w, 200, learning)
}

func trackQuery(r *http.Request) string {
	if id := r.URL.Query().Get("trackId"); id != "" {
		return "?trackId=" + url.QueryEscape(id)
	}
	return ""
}

func (s Server) subjects(w http.ResponseWriter, r *http.Request) {
	proxy(w, r, s.Learning, "GET", "/internal/subjects", nil, 200)
}
func (s Server) tracks(w http.ResponseWriter, r *http.Request) {
	proxy(w, r, s.Learning, "GET", "/internal/tracks", nil, 200)
}
func (s Server) createTrack(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid track")
		return
	}
	proxy(w, r, s.Learning, "POST", "/internal/tracks", body, 201)
}
func (s Server) curriculum(w http.ResponseWriter, r *http.Request) {
	proxy(w, r, s.Learning, "GET", "/internal/curriculum"+trackQuery(r), nil, 200)
}
func (s Server) progress(w http.ResponseWriter, r *http.Request) {
	proxy(w, r, s.Learning, "GET", "/internal/progress", nil, 200)
}

func (s Server) patchMe(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Timezone        *string `json:"timezone"`
		ProactivePaused *bool   `json:"proactivePaused"`
		QuietHours      *struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"quietHours"`
	}
	if platform.Decode(r, &in) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid preferences")
		return
	}
	if in.Timezone != nil {
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			platform.Error(w, 422, "validation_failed", "Invalid timezone")
			return
		}
	}
	if in.QuietHours != nil {
		if _, err := time.Parse("15:04", in.QuietHours.Start); err != nil {
			platform.Error(w, 422, "validation_failed", "Invalid quiet hours")
			return
		}
		if _, err := time.Parse("15:04", in.QuietHours.End); err != nil {
			platform.Error(w, 422, "validation_failed", "Invalid quiet hours")
			return
		}
	}
	var zone, start, end any
	if in.Timezone != nil {
		zone = *in.Timezone
	}
	if in.QuietHours != nil {
		start = in.QuietHours.Start
		end = in.QuietHours.End
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not update preferences")
		return
	}
	defer tx.Rollback()
	// Dispatch locks notification rows before the profile. Keep that order and
	// wake deferred invitations so a changed schedule takes effect immediately.
	if in.Timezone != nil || in.QuietHours != nil || in.ProactivePaused != nil {
		_, err = tx.ExecContext(r.Context(), `UPDATE opportunity.notification_outbox SET next_attempt_at=now(),updated_at=now()
 WHERE user_id=$1 AND state='pending' AND next_attempt_at>now()`, platform.User(r))
		if err != nil {
			platform.Error(w, 500, "database_error", "Could not update preferences")
			return
		}
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE experience.profiles SET timezone=COALESCE($2,timezone),proactive_paused=COALESCE($3,proactive_paused),
 quiet_start=COALESCE($4::time,quiet_start),quiet_end=COALESCE($5::time,quiet_end),updated_at=now() WHERE user_id=$1`, platform.User(r), zone, in.ProactivePaused, start, end)
	if err != nil || tx.Commit() != nil {
		platform.Error(w, 500, "database_error", "Could not update preferences")
		return
	}
	p, err := s.loadProfile(r)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not load preferences")
		return
	}
	platform.JSON(w, 200, p)
}

func (s Server) consent(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if len(key) == 0 || len(key) > 128 {
		platform.Error(w, 422, "idempotency_key_required", "Idempotency-Key is required")
		return
	}
	var in struct {
		Purpose       string `json:"purpose"`
		PolicyVersion string `json:"policyVersion"`
		Granted       bool   `json:"granted"`
	}
	if platform.Decode(r, &in) != nil || len(in.PolicyVersion) == 0 || len(in.PolicyVersion) > 32 {
		platform.Error(w, 422, "validation_failed", "Invalid consent")
		return
	}
	switch in.Purpose {
	case "calendar_context", "ios_calendar_context", "google_calendar_context", "location_context", "notifications", "shared_patterns":
	default:
		platform.Error(w, 422, "validation_failed", "Unknown consent purpose")
		return
	}
	var id string
	var granted bool
	var policyVersion string
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save consent")
		return
	}
	defer tx.Rollback()
	if err = lockProfile(r.Context(), tx, platform.User(r)); err != nil {
		platform.Error(w, 500, "database_error", "Could not save consent")
		return
	}
	err = tx.QueryRowContext(r.Context(), `INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,$3,$4,$5)
 ON CONFLICT(user_id,purpose,request_key) DO UPDATE SET request_key=experience.consents.request_key
 RETURNING id,granted,policy_version`, platform.User(r), in.Purpose, in.PolicyVersion, in.Granted, key).Scan(&id, &granted, &policyVersion)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not save consent")
		return
	}
	if granted != in.Granted || policyVersion != in.PolicyVersion {
		platform.Error(w, 409, "idempotency_conflict", "Request key was used for a different decision")
		return
	}
	if !granted && (in.Purpose == "location_context" || in.Purpose == "calendar_context" || in.Purpose == "ios_calendar_context" || in.Purpose == "google_calendar_context") {
		source := "ios_geofence"
		if in.Purpose == "calendar_context" {
			source = "calendar"
		} else if in.Purpose == "ios_calendar_context" {
			source = "ios_calendar"
		} else if in.Purpose == "google_calendar_context" {
			source = "google_calendar"
		}
		_, err = tx.ExecContext(r.Context(), `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source=$2 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, platform.User(r), source)
		if err != nil {
			platform.Error(w, 500, "database_error", "Could not revoke context")
			return
		}
	}
	if err = tx.Commit(); err != nil {
		platform.Error(w, 500, "database_error", "Could not save consent")
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id, "purpose": in.Purpose, "granted": granted})
}

func (s Server) device(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("installationId")
	if len(id) == 0 || len(id) > 128 {
		platform.Error(w, 422, "validation_failed", "Invalid installation")
		return
	}
	var in struct {
		PushToken                json.RawMessage `json:"pushToken"`
		NotificationPermission   string          `json:"notificationPermission"`
		LocationPermission       string          `json:"locationPermission"`
		CalendarPermission       string          `json:"calendarPermission"`
		GoogleCalendarPermission string          `json:"googleCalendarPermission"`
	}
	if platform.Decode(r, &in) != nil || (in.NotificationPermission != "" && in.NotificationPermission != "not_requested" && in.NotificationPermission != "granted" && in.NotificationPermission != "denied") ||
		(in.LocationPermission != "" && in.LocationPermission != "not_requested" && in.LocationPermission != "enabled" && in.LocationPermission != "denied") ||
		(in.CalendarPermission != "" && in.CalendarPermission != "not_requested" && in.CalendarPermission != "enabled" && in.CalendarPermission != "denied") ||
		(in.GoogleCalendarPermission != "" && in.GoogleCalendarPermission != "not_requested" && in.GoogleCalendarPermission != "enabled" && in.GoogleCalendarPermission != "denied") ||
		(in.NotificationPermission == "" && in.LocationPermission == "" && in.CalendarPermission == "" && in.GoogleCalendarPermission == "" && len(in.PushToken) == 0) {
		platform.Error(w, 422, "validation_failed", "Invalid device")
		return
	}
	var token *string
	if len(in.PushToken) != 0 && string(in.PushToken) != "null" {
		var value string
		if json.Unmarshal(in.PushToken, &value) != nil || len(value) > 255 ||
			!(strings.HasPrefix(value, "ExpoPushToken[") || strings.HasPrefix(value, "ExponentPushToken[")) || !strings.HasSuffix(value, "]") {
			platform.Error(w, 422, "validation_failed", "Invalid push token")
			return
		}
		token = &value
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not bind device")
		return
	}
	defer tx.Rollback()
	if err = lockProfile(r.Context(), tx, platform.User(r)); err != nil {
		platform.Error(w, 500, "database_error", "Could not bind device")
		return
	}
	if token != nil {
		if _, err = tx.ExecContext(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,2))", *token); err != nil {
			platform.Error(w, 500, "database_error", "Could not bind device")
			return
		}
		if _, err = tx.ExecContext(r.Context(), "UPDATE experience.devices SET push_token=NULL,notification_permission='denied' WHERE push_token=$1 AND user_id<>$2", *token, platform.User(r)); err != nil {
			platform.Error(w, 500, "database_error", "Could not bind device")
			return
		}
	}
	var stored string
	err = tx.QueryRowContext(r.Context(), `INSERT INTO experience.devices(installation_id,user_id,push_token,notification_permission,location_permission,calendar_permission,google_calendar_permission)
 VALUES($1,$2,$3,COALESCE(NULLIF($4,''),'not_requested'),COALESCE(NULLIF($5,''),'not_requested'),COALESCE(NULLIF($7,''),'not_requested'),COALESCE(NULLIF($8,''),'not_requested'))
 ON CONFLICT(installation_id) DO UPDATE SET push_token=CASE WHEN $6 THEN $3 ELSE experience.devices.push_token END,
 notification_permission=COALESCE(NULLIF($4,''),experience.devices.notification_permission),
 location_permission=COALESCE(NULLIF($5,''),experience.devices.location_permission),
 calendar_permission=COALESCE(NULLIF($7,''),experience.devices.calendar_permission),
 google_calendar_permission=COALESCE(NULLIF($8,''),experience.devices.google_calendar_permission),updated_at=now()
 WHERE experience.devices.user_id=excluded.user_id RETURNING installation_id`, id, platform.User(r), token, in.NotificationPermission, in.LocationPermission, len(in.PushToken) != 0, in.CalendarPermission, in.GoogleCalendarPermission).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		platform.Error(w, 409, "device_bound", "Device belongs to another account")
		return
	}
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not bind device")
		return
	}
	if in.CalendarPermission == "denied" {
		if err = cancelDeviceCalendarEvents(r.Context(), tx, platform.User(r), id, "ios_calendar", "ios_calendar_context", "ios:"); err != nil {
			platform.Error(w, 500, "database_error", "Could not revoke calendar access")
			return
		}
	}
	if in.GoogleCalendarPermission == "denied" {
		if err = cancelDeviceCalendarEvents(r.Context(), tx, platform.User(r), id, "google_calendar", "google_calendar_context", "google:"); err != nil {
			platform.Error(w, 500, "database_error", "Could not revoke calendar access")
			return
		}
	}
	if err = tx.Commit(); err != nil {
		platform.Error(w, 500, "database_error", "Could not bind device")
		return
	}
	platform.JSON(w, 200, map[string]string{"id": stored})
}

func (s Server) deleteDevice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("installationId")
	if len(id) == 0 || len(id) > 128 {
		platform.Error(w, 422, "validation_failed", "Invalid installation")
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	defer tx.Rollback()
	if err = lockProfile(r.Context(), tx, platform.User(r)); err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	_, err = tx.ExecContext(r.Context(), "DELETE FROM experience.devices WHERE installation_id=$1 AND user_id=$2", id, platform.User(r))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	var locationEnabled bool
	err = tx.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM experience.devices WHERE user_id=$1 AND location_permission='enabled')", platform.User(r)).Scan(&locationEnabled)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	if !locationEnabled {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 SELECT $1,'location_context',latest.policy_version,false,gen_random_uuid()::text FROM (
 SELECT policy_version,granted FROM experience.consents WHERE user_id=$1 AND purpose='location_context'
 ORDER BY decision_seq DESC LIMIT 1) latest WHERE latest.granted`, platform.User(r))
		if err == nil {
			_, err = tx.ExecContext(r.Context(), "UPDATE context.regions SET enabled=false WHERE user_id=$1 AND enabled", platform.User(r))
		}
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source='ios_geofence'
 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, platform.User(r))
		}
		if err != nil {
			platform.Error(w, 500, "database_error", "Could not remove device")
			return
		}
	}
	if err = cancelDeviceCalendarEvents(r.Context(), tx, platform.User(r), id, "ios_calendar", "ios_calendar_context", "ios:"); err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	if err = cancelDeviceCalendarEvents(r.Context(), tx, platform.User(r), id, "google_calendar", "google_calendar_context", "google:"); err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	if err = tx.Commit(); err != nil {
		platform.Error(w, 500, "database_error", "Could not remove device")
		return
	}
	platform.JSON(w, 200, map[string]bool{"removed": true})
}

func cancelDeviceCalendarEvents(ctx context.Context, tx *sql.Tx, user, installation, source, purpose, prefix string) error {
	_, err := tx.ExecContext(ctx, `WITH canceled AS (
 UPDATE context.events SET canceled_at=now() WHERE user_id=$1 AND source=$2
 AND source_resource_id=$3 AND canceled_at IS NULL RETURNING id
 ) INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING`, user, source, prefix+installation)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 SELECT $1,$2,latest.policy_version,false,gen_random_uuid()::text FROM (
 SELECT policy_version,granted FROM experience.consents WHERE user_id=$1 AND purpose=$2
 ORDER BY decision_seq DESC LIMIT 1) latest WHERE latest.granted AND NOT EXISTS (
 SELECT 1 FROM experience.devices WHERE user_id=$1 AND
 CASE $3 WHEN 'ios_calendar' THEN calendar_permission WHEN 'google_calendar' THEN google_calendar_permission ELSE '' END='enabled')`, user, purpose, source)
	return err
}

func (s Server) notificationStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("installationId")
	if len(id) == 0 || len(id) > 128 {
		platform.Error(w, 422, "validation_failed", "Invalid installation")
		return
	}
	var enabled bool
	err := s.DB.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM experience.devices d WHERE d.user_id=$1 AND d.installation_id=$2
 AND d.notification_permission='granted' AND d.push_token IS NOT NULL) AND
 COALESCE((SELECT c.granted FROM experience.consents c WHERE c.user_id=$1 AND c.purpose='notifications' ORDER BY c.decision_seq DESC LIMIT 1),false)`, platform.User(r), id).Scan(&enabled)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not check notifications")
		return
	}
	platform.JSON(w, 200, map[string]bool{"available": platform.Env("PUSH_ENABLED", "false") == "true", "enabled": enabled})
}

func (s Server) regions(w http.ResponseWriter, r *http.Request) {
	proxy(w, r, s.Context, "GET", "/internal/regions", nil, 200)
}
func (s Server) saveRegion(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid place")
		return
	}
	proxy(w, r, s.Context, "POST", "/internal/regions", body, 201)
}
func (s Server) deleteRegions(w http.ResponseWriter, r *http.Request) {
	if owner := r.URL.Query().Get("ownerId"); owner != "" && owner != platform.User(r) {
		platform.Error(w, 409, "account_changed", "Saved place belongs to another account")
		return
	}
	proxy(w, r, s.Context, "DELETE", "/internal/regions", nil, 200)
}
func (s Server) location(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid signal")
		return
	}
	proxy(w, r, s.Context, "POST", "/internal/signals/location", body, 202)
}
func (s Server) calendarWindows(w http.ResponseWriter, r *http.Request, provider string) {
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid availability")
		return
	}
	proxy(w, r, s.Context, "POST", "/internal/calendar/"+provider+"/windows", body, 202)
}
func (s Server) deleteCalendarWindows(w http.ResponseWriter, r *http.Request, provider string) {
	proxy(w, r, s.Context, "DELETE", "/internal/calendar/"+provider+"/windows?installationId="+url.QueryEscape(r.URL.Query().Get("installationId")), nil, 200)
}
func (s Server) offers(w http.ResponseWriter, r *http.Request) {
	allowed, err := s.proactiveAllowed(r, "")
	if err != nil {
		platform.Error(w, 500, "eligibility_error", "Could not check preferences")
		return
	}
	if !allowed {
		platform.JSON(w, 200, map[string]any{"items": []any{}, "nextCursor": nil})
		return
	}
	proxy(w, r, s.Opportunity, "GET", "/internal/offers", nil, 200)
}
func (s Server) offerAction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid opportunity")
		return
	}
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid action")
		return
	}
	allowed, err := s.offerContextAllowed(r, id)
	if err != nil {
		remoteError(w, err)
		return
	}
	if !allowed {
		platform.Error(w, 403, "opportunity_unavailable", "Context opportunity is paused or revoked")
		return
	}
	proxy(w, r, s.Opportunity, "POST", "/internal/offers/"+id+"/actions", body, 200)
}

func (s Server) application(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid opportunity")
		return
	}
	var in struct {
		Outcome string `json:"outcome"`
	}
	if platform.Decode(r, &in) != nil || (in.Outcome != "applied" && in.Outcome != "tried" && in.Outcome != "missed") {
		platform.Error(w, 422, "validation_failed", "Invalid application report")
		return
	}
	var offer struct {
		TrackID        string    `json:"trackId"`
		NodeID         string    `json:"nodeId"`
		ContextEventID string    `json:"contextEventId"`
		Status         string    `json:"status"`
		ValidUntil     time.Time `json:"validUntil"`
	}
	if err := s.Opportunity.Do(r.Context(), platform.User(r), "GET", "/internal/offers/"+id, nil, &offer); err != nil {
		remoteError(w, err)
		return
	}
	if (offer.Status != "ready" && offer.Status != "completed") || !offer.ValidUntil.After(time.Now()) {
		platform.Error(w, 409, "opportunity_unavailable", "Opportunity is no longer current")
		return
	}
	var event struct {
		Canceled   bool      `json:"canceled"`
		Source     string    `json:"source"`
		ValidUntil time.Time `json:"validUntil"`
	}
	if err := s.Context.Do(r.Context(), platform.User(r), "GET", "/internal/events/"+offer.ContextEventID, nil, &event); err != nil {
		remoteError(w, err)
		return
	}
	if event.Canceled || !event.ValidUntil.After(time.Now()) {
		platform.Error(w, 409, "opportunity_unavailable", "Context is no longer current")
		return
	}
	allowed, err := s.proactiveAllowed(r, event.Source)
	if err != nil {
		platform.Error(w, 500, "eligibility_error", "Could not check preferences")
		return
	}
	if !allowed {
		platform.Error(w, 403, "opportunity_unavailable", "Context opportunity is paused or revoked")
		return
	}
	var result any
	err = s.Learning.Do(r.Context(), platform.User(r), "POST", "/internal/applications", map[string]string{"opportunityId": id, "trackId": offer.TrackID, "nodeId": offer.NodeID, "outcome": in.Outcome}, &result)
	if err != nil {
		remoteError(w, err)
		return
	}
	if err = s.Opportunity.Do(r.Context(), platform.User(r), "POST", "/internal/offers/"+id+"/complete", nil, nil); err != nil {
		remoteError(w, err)
		return
	}
	platform.JSON(w, 201, result)
}

func (s Server) startSession(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if len(key) == 0 || len(key) > 128 {
		platform.Error(w, 422, "idempotency_key_required", "Idempotency-Key is required")
		return
	}
	var in struct {
		Source struct {
			Type    string `json:"type"`
			NodeID  string `json:"nodeId"`
			ID      string `json:"id"`
			TrackID string `json:"trackId"`
		} `json:"source"`
		Mode string `json:"mode"`
	}
	if platform.Decode(r, &in) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid session")
		return
	}
	body := map[string]string{"trackId": in.Source.TrackID, "nodeId": in.Source.NodeID, "sourceType": in.Source.Type, "sourceId": in.Source.NodeID, "mode": in.Mode, "idempotencyKey": key}
	if in.Source.Type == "opportunity" {
		if platform.CheckUUID(in.Source.ID) != nil {
			platform.Error(w, 422, "validation_failed", "Invalid opportunity")
			return
		}
		var offer struct {
			TrackID        string    `json:"trackId"`
			NodeID         string    `json:"nodeId"`
			Status         string    `json:"status"`
			ValidUntil     time.Time `json:"validUntil"`
			ContextEventID string    `json:"contextEventId"`
		}
		if err := s.Opportunity.Do(r.Context(), platform.User(r), "GET", "/internal/offers/"+in.Source.ID, nil, &offer); err != nil {
			remoteError(w, err)
			return
		}
		if offer.Status != "ready" || offer.ValidUntil.Before(time.Now()) {
			platform.Error(w, 409, "opportunity_unavailable", "Opportunity is no longer ready")
			return
		}
		var event struct {
			Canceled   bool      `json:"canceled"`
			Source     string    `json:"source"`
			ValidUntil time.Time `json:"validUntil"`
		}
		if err := s.Context.Do(r.Context(), platform.User(r), "GET", "/internal/events/"+offer.ContextEventID, nil, &event); err != nil {
			remoteError(w, err)
			return
		}
		if event.Canceled || !event.ValidUntil.After(time.Now()) {
			platform.Error(w, 409, "opportunity_unavailable", "Context is no longer current")
			return
		}
		allowed, err := s.proactiveAllowed(r, event.Source)
		if err != nil {
			platform.Error(w, 500, "eligibility_error", "Could not check preferences")
			return
		}
		if !allowed {
			platform.Error(w, 403, "opportunity_unavailable", "Context opportunity is paused or revoked")
			return
		}
		body["trackId"] = offer.TrackID
		body["nodeId"] = offer.NodeID
		body["sourceId"] = in.Source.ID
	} else if in.Source.Type != "curriculum" {
		platform.Error(w, 422, "validation_failed", "Unsupported session source")
		return
	}
	if body["nodeId"] == "" {
		platform.Error(w, 422, "validation_failed", "Activity is required")
		return
	}
	proxy(w, r, s.Learning, "POST", "/internal/sessions", body, 201)
}

func (s Server) offerContextAllowed(r *http.Request, id string) (bool, error) {
	var offer struct {
		ContextEventID string `json:"contextEventId"`
	}
	if err := s.Opportunity.Do(r.Context(), platform.User(r), "GET", "/internal/offers/"+id, nil, &offer); err != nil {
		return false, err
	}
	var event struct {
		Source     string    `json:"source"`
		Canceled   bool      `json:"canceled"`
		ValidUntil time.Time `json:"validUntil"`
	}
	if err := s.Context.Do(r.Context(), platform.User(r), "GET", "/internal/events/"+offer.ContextEventID, nil, &event); err != nil {
		return false, err
	}
	if event.Canceled || !event.ValidUntil.After(time.Now()) {
		return false, nil
	}
	if event.Source == "ios_calendar" || event.Source == "google_calendar" {
		var sharedGap bool
		if err := s.DB.QueryRowContext(r.Context(), "SELECT context.calendar_sources_agree($1,$2)", offer.ContextEventID, platform.User(r)).Scan(&sharedGap); err != nil {
			return false, err
		}
		if !sharedGap {
			return false, nil
		}
	}
	return s.proactiveAllowed(r, event.Source)
}
func (s Server) turn(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid session")
		return
	}
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid turn")
		return
	}
	proxy(w, r, s.Learning, "POST", "/internal/sessions/"+id+"/turns", body, 200)
}
func (s Server) voice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid session")
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
	if err != nil || len(data) > 1<<20 {
		platform.Error(w, 413, "too_large", "Voice request is too large")
		return
	}
	var body json.RawMessage
	if !json.Valid(data) {
		platform.Error(w, 422, "validation_failed", "Invalid voice request")
		return
	}
	body = data
	path := "/internal/sessions/" + id + "/speech"
	if r.URL.Path == "/v1/sessions/"+id+"/transcribe" {
		path = "/internal/sessions/" + id + "/transcribe"
	}
	proxy(w, r, s.Learning, "POST", path, body, 200)
}
func (s Server) realtime(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid session")
		return
	}
	var result struct {
		ClientSecret string `json:"clientSecret"`
	}
	if err := s.Learning.Do(r.Context(), platform.User(r), "POST", "/internal/sessions/"+id+"/realtime", nil, &result); err != nil {
		remoteError(w, err)
		return
	}
	platform.JSON(w, 200, result)
}
func (s Server) finish(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if platform.CheckUUID(id) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid session")
		return
	}
	var body json.RawMessage
	if platform.Decode(r, &body) != nil {
		platform.Error(w, 422, "validation_failed", "Invalid finish")
		return
	}
	var result struct {
		ID         string `json:"id"`
		State      string `json:"state"`
		SourceType string `json:"sourceType"`
		SourceID   string `json:"sourceId"`
	}
	if err := s.Learning.Do(r.Context(), platform.User(r), "POST", "/internal/sessions/"+id+"/finish", body, &result); err != nil {
		remoteError(w, err)
		return
	}
	if result.ID != id || (result.State != "completed" && result.State != "abandoned") {
		platform.Error(w, 503, "service_unavailable", "Could not confirm session state")
		return
	}
	if result.State == "completed" && result.SourceType == "opportunity" {
		if platform.CheckUUID(result.SourceID) != nil {
			platform.Error(w, 503, "service_unavailable", "Could not confirm opportunity")
			return
		}
		err := s.Opportunity.Do(r.Context(), platform.User(r), "POST", "/internal/offers/"+result.SourceID+"/complete", nil, nil)
		var remote platform.RemoteError
		if err != nil && (!errors.As(err, &remote) || remote.Status != http.StatusConflict) {
			remoteError(w, err)
			return
		}
	}
	platform.JSON(w, 202, map[string]string{"id": result.ID, "state": result.State})
}
func (s Server) billing(w http.ResponseWriter, r *http.Request) {
	var summary map[string]any
	if err := s.Learning.Do(r.Context(), platform.User(r), "GET", "/internal/summary", nil, &summary); err != nil {
		remoteError(w, err)
		return
	}
	allowance, _ := summary["allowance"].(map[string]any)
	platform.JSON(w, 200, map[string]any{"entitlement": nil, "freeRemaining": allowance["freeRemaining"], "paidSecondsRemaining": 0})
}

func (s Server) locationEligible(w http.ResponseWriter, r *http.Request) {
	installation := r.URL.Query().Get("installationId")
	var count int
	err := s.DB.QueryRowContext(r.Context(), `SELECT count(*) FROM experience.devices d WHERE d.user_id=$1 AND d.installation_id=$2
 AND d.location_permission='enabled'`, platform.User(r), installation).Scan(&count)
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not check eligibility")
		return
	}
	var granted bool
	_ = s.DB.QueryRowContext(r.Context(), "SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose='location_context' ORDER BY decision_seq DESC LIMIT 1", platform.User(r)).Scan(&granted)
	platform.JSON(w, 200, map[string]bool{"allowed": count > 0 && granted})
}

func (s Server) opportunityEligible(w http.ResponseWriter, r *http.Request) {
	// Quiet hours gate outbound notifications, not an offer the learner chooses to open.
	allowed, err := s.proactiveAllowed(r, r.URL.Query().Get("source"))
	if err != nil {
		platform.Error(w, 500, "database_error", "Could not check preferences")
		return
	}
	platform.JSON(w, 200, map[string]bool{"allowed": allowed})
}

func (s Server) proactiveAllowed(r *http.Request, source string) (bool, error) {
	p, err := s.loadProfile(r)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if p.ProactivePaused {
		return false, nil
	}
	purpose := ""
	switch source {
	case "ios_geofence":
		purpose = "location_context"
	case "calendar":
		purpose = "calendar_context"
	case "ios_calendar":
		purpose = "ios_calendar_context"
	case "google_calendar":
		purpose = "google_calendar_context"
	case "":
	default:
		return false, nil
	}
	var granted bool
	if purpose == "" {
		err = s.DB.QueryRowContext(r.Context(), `SELECT COALESCE(bool_or(granted),false) FROM (
 SELECT DISTINCT ON (purpose) granted FROM experience.consents
 WHERE user_id=$1 AND purpose IN ('location_context','calendar_context','ios_calendar_context','google_calendar_context') ORDER BY purpose,decision_seq DESC
 ) latest`, platform.User(r)).Scan(&granted)
	} else {
		err = s.DB.QueryRowContext(r.Context(), "SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose=$2 ORDER BY decision_seq DESC LIMIT 1", platform.User(r), purpose).Scan(&granted)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return granted, nil
}

func inQuiet(now, start, end string) bool {
	if start == end {
		return false
	}
	if start < end {
		return now >= start && now < end
	}
	return now >= start || now < end
}

func proxy(w http.ResponseWriter, r *http.Request, c platform.Client, method, path string, input any, status int) {
	var result any
	if err := c.Do(r.Context(), platform.User(r), method, path, input, &result); err != nil {
		remoteError(w, err)
		return
	}
	platform.JSON(w, status, result)
}

func remoteError(w http.ResponseWriter, err error) {
	var remote platform.RemoteError
	if errors.As(err, &remote) {
		if remote.Status >= 400 && remote.Status < 500 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(remote.Status)
			_, _ = w.Write([]byte(remote.Body))
			return
		}
	}
	platform.Error(w, 503, "service_unavailable", "A service is temporarily unavailable")
}
