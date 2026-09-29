package experience

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"learning-orchestrator/backend/internal/platform"
)

// Account export is a privacy boundary: every query is owner-scoped, and
// credentials, provider tickets, request keys, and internal work queues stay out.
var exportSections = []struct{ name, query string }{
	{"profile", `SELECT COALESCE((SELECT to_jsonb(p) FROM experience.profiles p WHERE p.user_id=$1),'null'::jsonb)`},
	{"devices", `SELECT COALESCE(jsonb_agg(jsonb_build_object('installationId',d.installation_id,'notificationPermission',d.notification_permission,'locationPermission',d.location_permission,'iosCalendarPermission',d.calendar_permission,'googleCalendarPermission',d.google_calendar_permission,'updatedAt',d.updated_at)),'[]'::jsonb) FROM experience.devices d WHERE d.user_id=$1`},
	{"consents", `SELECT COALESCE(jsonb_agg(jsonb_build_object('id',c.id,'purpose',c.purpose,'policyVersion',c.policy_version,'granted',c.granted,'createdAt',c.created_at)),'[]'::jsonb) FROM experience.consents c WHERE c.user_id=$1`},
	{"tracks", `SELECT COALESCE(jsonb_agg(to_jsonb(t)),'[]'::jsonb) FROM learning.tracks t WHERE t.user_id=$1`},
	{"courses", `SELECT COALESCE(jsonb_agg(to_jsonb(c)),'[]'::jsonb) FROM learning.track_content c JOIN learning.tracks t ON t.id=c.track_id WHERE t.user_id=$1`},
	{"progress", `SELECT COALESCE(jsonb_agg(to_jsonb(p)),'[]'::jsonb) FROM learning.progress p JOIN learning.tracks t ON t.id=p.track_id WHERE t.user_id=$1`},
	{"sessions", `SELECT COALESCE(jsonb_agg(to_jsonb(s)),'[]'::jsonb) FROM learning.sessions s WHERE s.user_id=$1`},
	{"turns", `SELECT COALESCE(jsonb_agg(to_jsonb(n)),'[]'::jsonb) FROM learning.turns n JOIN learning.sessions s ON s.id=n.session_id WHERE s.user_id=$1`},
	{"evidence", `SELECT COALESCE(jsonb_agg(to_jsonb(e)),'[]'::jsonb) FROM learning.evidence e WHERE e.user_id=$1`},
	{"skillState", `SELECT COALESCE(jsonb_agg(to_jsonb(s)),'[]'::jsonb) FROM learning.skill_state s JOIN learning.tracks t ON t.id=s.track_id WHERE t.user_id=$1`},
	{"courseGenerationUsage", `SELECT COALESCE(jsonb_agg(to_jsonb(q)),'[]'::jsonb) FROM learning.course_generation_quota q WHERE q.user_id=$1`},
	{"savedPlaces", `SELECT COALESCE(jsonb_agg(to_jsonb(r)),'[]'::jsonb) FROM context.regions r WHERE r.user_id=$1`},
	{"contextEvents", `SELECT COALESCE(jsonb_agg(to_jsonb(e)),'[]'::jsonb) FROM context.events e WHERE e.user_id=$1`},
	{"opportunities", `SELECT COALESCE(jsonb_agg(to_jsonb(o)),'[]'::jsonb) FROM opportunity.offers o WHERE o.user_id=$1`},
	{"notifications", `SELECT COALESCE(jsonb_agg(jsonb_build_object('offerId',n.offer_id,'state',n.state,'sentAt',n.sent_at,'receiptStatus',n.receipt_status,'receiptCheckedAt',n.receipt_checked_at)),'[]'::jsonb)
 FROM opportunity.notification_outbox n WHERE n.user_id=$1`},
}

func (s Server) exportData(w http.ResponseWriter, r *http.Request) {
	tx, err := s.DB.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		platform.Error(w, 500, "export_failed", "Could not prepare your data export")
		return
	}
	defer tx.Rollback()
	data := make(map[string]json.RawMessage, len(exportSections))
	for _, section := range exportSections {
		var rows json.RawMessage
		if err = tx.QueryRowContext(r.Context(), section.query, platform.User(r)).Scan(&rows); err != nil {
			platform.Error(w, 500, "export_failed", "Could not prepare your data export")
			return
		}
		data[section.name] = rows
	}
	if err = tx.Commit(); err != nil {
		platform.Error(w, 500, "export_failed", "Could not prepare your data export")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	platform.JSON(w, 200, map[string]any{"schemaVersion": 1, "exportedAt": time.Now().UTC(), "data": data})
}
