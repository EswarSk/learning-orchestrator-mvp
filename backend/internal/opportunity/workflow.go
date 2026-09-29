package opportunity

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"learning-orchestrator/backend/internal/contracts"
	"learning-orchestrator/backend/internal/platform"
)

func Workflow(ctx workflow.Context, input contracts.ContextTrigger) error {
	var wait time.Duration
	if !input.EvaluateAt.IsZero() {
		wait = input.EvaluateAt.Sub(workflow.Now(ctx))
	} else {
		delay := input.DelaySeconds
		if delay < 1 || delay > 3600 {
			delay = 120
		}
		wait = time.Duration(delay) * time.Second
	}
	if wait > 0 {
		if err := workflow.Sleep(ctx, wait); err != nil {
			return err
		}
	}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: 15 * time.Second, RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 5}})
	return workflow.ExecuteActivity(ctx, "EvaluateOpportunity", input).Get(ctx, nil)
}

type Activities struct {
	DB         *sql.DB
	Context    platform.Client
	Experience platform.Client
	Learning   platform.Client
}
type contextEvent struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	TrackID    string    `json:"trackId"`
	Category   string    `json:"category"`
	ObservedAt time.Time `json:"observedAt"`
	ValidUntil time.Time `json:"validUntil"`
	Canceled   bool      `json:"canceled"`
}

// ponytail: Foreground calendar scans are trusted for five minutes; background sync is needed for longer-lived offers.
const foregroundCalendarFreshness = 5 * time.Minute

func foregroundCalendar(source string) bool {
	return source == "ios_calendar" || source == "google_calendar"
}

type candidate struct {
	TrackID    string `json:"trackId"`
	NodeID     string `json:"nodeId"`
	Title      string `json:"title"`
	ReasonCode string `json:"reasonCode"`
}
type candidateResponse struct {
	Candidate *candidate `json:"candidate"`
}
type eligibility struct {
	Allowed bool `json:"allowed"`
}

func (a Activities) Evaluate(ctx context.Context, in contracts.ContextTrigger) error {
	if platform.CheckUUID(in.EventID) != nil || in.UserID == "" {
		return temporal.NewNonRetryableApplicationError("invalid trigger", "invalid_trigger", nil)
	}
	var event contextEvent
	if err := a.Context.Do(ctx, in.UserID, "GET", "/internal/events/"+in.EventID, nil, &event); err != nil {
		return err
	}
	if event.Canceled || !event.ValidUntil.After(time.Now()) || (foregroundCalendar(event.Source) && !freshForegroundCalendar(event.ObservedAt, time.Now())) {
		return nil
	}
	var allowed eligibility
	if err := a.Experience.Do(ctx, in.UserID, "GET", "/internal/opportunity-eligible?source="+url.QueryEscape(event.Source), nil, &allowed); err != nil {
		return err
	}
	if !allowed.Allowed {
		return nil
	}
	var result candidateResponse
	candidatePath := "/internal/candidates?category=" + url.QueryEscape(event.Category)
	if event.TrackID != "" {
		candidatePath += "&trackId=" + url.QueryEscape(event.TrackID)
	}
	if err := a.Learning.Do(ctx, in.UserID, "GET", candidatePath, nil, &result); err != nil {
		return err
	}
	if result.Candidate == nil {
		return nil
	}
	// One candidate per context; Learning ranks due reviews across tracks.
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", in.UserID); err != nil {
		return err
	}
	// Consent changes and preference updates serialize on this profile row.
	// The earlier service check can be stale after the remote candidate lookup.
	var paused bool
	err = tx.QueryRowContext(ctx, `SELECT proactive_paused FROM experience.profiles WHERE user_id=$1 FOR UPDATE`, in.UserID).Scan(&paused)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if paused {
		return nil
	}
	var active, recent int
	var last sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT
 count(*) FILTER (WHERE status IN ('ready','offered','scheduled') AND valid_until>now()),
 count(*) FILTER (WHERE created_at>now()-interval '24 hours' AND status NOT IN ('canceled','suppressed')),
 max(created_at) FILTER (WHERE status NOT IN ('canceled','suppressed'))
 FROM opportunity.offers WHERE user_id=$1`, in.UserID).Scan(&active, &recent, &last)
	if err != nil {
		return err
	}
	if !mayOffer(active, recent, last, time.Now()) {
		return nil
	}
	// Serialize the final validity check with a geofence exit. If cancellation
	// wins first, there is no offer; if this insert wins, its cancel outbox
	// subsequently closes the offer.
	var current bool
	var source string
	var observedAt time.Time
	err = tx.QueryRowContext(ctx, `SELECT canceled_at IS NULL AND valid_until>now(),source,observed_at FROM context.events
	 WHERE id=$1 AND user_id=$2 FOR UPDATE`, in.EventID, in.UserID).Scan(&current, &source, &observedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !current || (foregroundCalendar(source) && !freshForegroundCalendar(observedAt, time.Now())) {
		return nil
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
	default:
		return nil
	}
	var granted bool
	err = tx.QueryRowContext(ctx, `SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose=$2
	 ORDER BY decision_seq DESC LIMIT 1`, in.UserID, purpose).Scan(&granted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !granted {
		return nil
	}
	if foregroundCalendar(source) {
		var sharedGap bool
		if err = tx.QueryRowContext(ctx, "SELECT context.calendar_sources_agree($1,$2)", in.EventID, in.UserID).Scan(&sharedGap); err != nil {
			return err
		}
		if !sharedGap {
			return nil
		}
	}
	var id string
	reason := result.Candidate.ReasonCode
	if reason != "saved_practice_place" && reason != "due_review_at_saved_place" && reason != "calendar_availability" && reason != "review_due_in_calendar_gap" {
		reason = "approved_context_match"
	}
	validUntil := event.ValidUntil
	if foregroundCalendar(source) && observedAt.Add(foregroundCalendarFreshness).Before(validUntil) {
		validUntil = observedAt.Add(foregroundCalendarFreshness)
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(context_event_id) DO UPDATE SET updated_at=opportunity.offers.updated_at RETURNING id`,
		in.UserID, result.Candidate.TrackID, result.Candidate.NodeID, in.EventID, result.Candidate.Title, reason, validUntil).Scan(&id)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO opportunity.notification_outbox(offer_id,user_id) VALUES($1,$2)
 ON CONFLICT(offer_id) DO NOTHING`, id, in.UserID); err != nil {
		return err
	}
	return tx.Commit()
}

func freshForegroundCalendar(observedAt, now time.Time) bool {
	return !observedAt.IsZero() && !observedAt.After(now.Add(time.Minute)) && now.Sub(observedAt) < foregroundCalendarFreshness
}

func mayOffer(active, recent int, last sql.NullTime, now time.Time) bool {
	return active == 0 && recent < 2 && (!last.Valid || now.Sub(last.Time) >= 3*time.Hour)
}
