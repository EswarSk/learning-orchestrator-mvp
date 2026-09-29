package opportunity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

// DispatchNotifications sends only opted-in, still-valid invitations. The row locks
// serialize a send with offer cancellation and context invalidation.
func DispatchNotifications(ctx context.Context, db *sql.DB) {
	client := &http.Client{Timeout: 7 * time.Second}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if err := dispatchNotification(ctx, db, client, expoPushURL); err != nil && ctx.Err() == nil {
			slog.Error("notification dispatch failed", "error", err)
		}
		if err := checkReceipts(ctx, db, client, expoReceiptsURL); err != nil && ctx.Err() == nil {
			slog.Error("push receipt check failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func dispatchNotification(ctx context.Context, db *sql.DB, client *http.Client, endpoint string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var offerID, userID, installationID, token, status, source, timezone, quietStart, quietEnd string
	var canceled sql.NullTime
	var validUntil, eventValidUntil time.Time
	var paused, notificationConsent, contextConsent bool
	var sentToday, attempts int
	err = tx.QueryRowContext(ctx, `SELECT n.offer_id::text,n.user_id,o.status,o.valid_until,e.source,e.canceled_at,e.valid_until,
 p.timezone,to_char(p.quiet_start,'HH24:MI'),to_char(p.quiet_end,'HH24:MI'),p.proactive_paused,n.attempts
 FROM opportunity.notification_outbox n
 JOIN opportunity.offers o ON o.id=n.offer_id AND o.user_id=n.user_id
 JOIN context.events e ON e.id=o.context_event_id AND e.user_id=n.user_id
 JOIN experience.profiles p ON p.user_id=n.user_id
 WHERE n.state='pending' AND n.next_attempt_at<=now()
 ORDER BY n.next_attempt_at LIMIT 1 FOR UPDATE OF n,o,e,p SKIP LOCKED`).Scan(
		&offerID, &userID, &status, &validUntil, &source, &canceled, &eventValidUntil, &timezone,
		&quietStart, &quietEnd, &paused, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	// Read consent and token after locking the profile; enrollment/revocation takes
	// the same lock, so a queued send cannot use a stale permission snapshot.
	err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT c.granted FROM experience.consents c WHERE c.user_id=$1 AND c.purpose='notifications' ORDER BY c.decision_seq DESC LIMIT 1),false),
 COALESCE((SELECT c.granted FROM experience.consents c WHERE c.user_id=$1 AND c.purpose=CASE $2 WHEN 'ios_geofence' THEN 'location_context' WHEN 'calendar' THEN 'calendar_context' WHEN 'ios_calendar' THEN 'ios_calendar_context' WHEN 'google_calendar' THEN 'google_calendar_context' ELSE '' END ORDER BY c.decision_seq DESC LIMIT 1),false),
 (SELECT count(*) FROM opportunity.notification_outbox sent WHERE sent.user_id=$1 AND sent.state='sent'
 AND (sent.sent_at AT TIME ZONE $3)::date=(now() AT TIME ZONE $3)::date)`, userID, source, timezone).Scan(&notificationConsent, &contextConsent, &sentToday)
	if err != nil {
		return err
	}
	err = tx.QueryRowContext(ctx, `SELECT installation_id,push_token FROM experience.devices WHERE user_id=$1
 AND notification_permission='granted' AND push_token IS NOT NULL ORDER BY updated_at DESC LIMIT 1 FOR UPDATE`, userID).Scan(&installationID, &token)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := time.Now()
	zone, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}
	quiet := quietAt(now, zone, quietStart, quietEnd)
	if status != "ready" || canceled.Valid || !validUntil.After(now) || !eventValidUntil.After(now) ||
		// Foreground-only EventKit refresh cannot justify an unattended push after a calendar edit.
		(source != "ios_geofence" && source != "calendar") || paused || !notificationConsent || !contextConsent ||
		sentToday >= 1 || token == "" {
		_, err = tx.ExecContext(ctx, "UPDATE opportunity.notification_outbox SET state='suppressed',updated_at=now() WHERE offer_id=$1", offerID)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	if quiet {
		wake := now.Truncate(time.Minute).Add(time.Minute)
		for quietAt(wake, zone, quietStart, quietEnd) && wake.Before(validUntil) && wake.Before(eventValidUntil) {
			wake = wake.Add(time.Minute)
		}
		if !validUntil.After(wake) || !eventValidUntil.After(wake) {
			_, err = tx.ExecContext(ctx, "UPDATE opportunity.notification_outbox SET state='suppressed',updated_at=now() WHERE offer_id=$1", offerID)
		} else {
			_, err = tx.ExecContext(ctx, "UPDATE opportunity.notification_outbox SET next_attempt_at=$2,updated_at=now() WHERE offer_id=$1", offerID, wake)
		}
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	if !strings.HasPrefix(token, "ExpoPushToken[") && !strings.HasPrefix(token, "ExponentPushToken[") {
		if _, err = tx.ExecContext(ctx, "UPDATE opportunity.notification_outbox SET state='failed',updated_at=now() WHERE offer_id=$1", offerID); err != nil {
			return err
		}
		return tx.Commit()
	}
	message, _ := json.Marshal(map[string]any{
		"to": token, "title": "A learning moment is ready", "body": "Open the app when you have a moment to practice.",
		"data": map[string]string{"offerId": offerID}, "ttl": int(time.Until(validUntil).Seconds()),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, sendErr := client.Do(req)
	if sendErr == nil {
		defer res.Body.Close()
	}
	var ticket struct {
		Data struct {
			Status  string `json:"status"`
			ID      string `json:"id"`
			Details struct {
				Error string `json:"error"`
			} `json:"details"`
		} `json:"data"`
	}
	if sendErr == nil && res.StatusCode == http.StatusOK {
		sendErr = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&ticket)
	}
	if sendErr == nil && res.StatusCode == http.StatusOK && ticket.Data.Status == "ok" && ticket.Data.ID != "" {
		tokenHash := sha256.Sum256([]byte(token))
		_, err = tx.ExecContext(ctx, `UPDATE opportunity.notification_outbox SET state='sent',attempts=attempts+1,
 provider_ticket=$2,device_installation_id=$3,push_token_hash=$4,receipt_status='pending',
 receipt_next_check_at=now()+interval '15 minutes',sent_at=now(),updated_at=now() WHERE offer_id=$1`,
			offerID, ticket.Data.ID, installationID, fmt.Sprintf("%x", tokenHash))
	} else {
		permanent := res != nil && res.StatusCode >= 400 && res.StatusCode < 500 && res.StatusCode != 429
		permanent = permanent || ticket.Data.Status == "error" || attempts >= 3 || !validUntil.After(now.Add(time.Minute))
		state, delay := "pending", time.Duration(1<<min(attempts, 4))*time.Minute
		if permanent {
			state = "failed"
		}
		_, err = tx.ExecContext(ctx, `UPDATE opportunity.notification_outbox SET state=$2,attempts=attempts+1,
 next_attempt_at=now()+$3::interval,updated_at=now() WHERE offer_id=$1`, offerID, state, fmt.Sprintf("%d seconds", int(delay.Seconds())))
		if ticket.Data.Details.Error == "DeviceNotRegistered" {
			if _, err = tx.ExecContext(ctx, "UPDATE experience.devices SET push_token=NULL,notification_permission='denied' WHERE user_id=$1 AND push_token=$2", userID, token); err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func quietAt(now time.Time, zone *time.Location, start, end string) bool {
	if start == end {
		return false
	}
	local := now.In(zone).Format("15:04")
	if start < end {
		return local >= start && local < end
	}
	return local >= start || local < end
}
