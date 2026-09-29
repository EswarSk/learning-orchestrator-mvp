package opportunity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const expoReceiptsURL = "https://exp.host/--/api/v2/push/getReceipts"

type dueReceipt struct {
	offerID, userID, ticket, installationID, tokenHash string
	attempts                                           int
}

func checkReceipts(ctx context.Context, db *sql.DB, client *http.Client, endpoint string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT offer_id::text,user_id,provider_ticket,
 COALESCE(device_installation_id,''),COALESCE(push_token_hash,''),receipt_attempts
 FROM opportunity.notification_outbox WHERE state='sent' AND receipt_status='pending'
 AND receipt_next_check_at<=now() ORDER BY receipt_next_check_at LIMIT 100 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	due := []dueReceipt{}
	for rows.Next() {
		var item dueReceipt
		if err = rows.Scan(&item.offerID, &item.userID, &item.ticket, &item.installationID, &item.tokenHash, &item.attempts); err != nil {
			rows.Close()
			return err
		}
		due = append(due, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if len(due) == 0 {
		return nil
	}
	ids := make([]string, len(due))
	for i, item := range due {
		ids[i] = item.ticket
	}
	body, _ := json.Marshal(map[string][]string{"ids": ids})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, requestErr := client.Do(req)
	var payload struct {
		Data map[string]struct {
			Status  string `json:"status"`
			Details struct {
				Error string `json:"error"`
			} `json:"details"`
		} `json:"data"`
	}
	if requestErr == nil {
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			requestErr = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&payload)
		}
		if res.StatusCode != http.StatusOK {
			requestErr = fmt.Errorf("receipt service HTTP %d", res.StatusCode)
		}
	}
	type failure struct{ offerID, state, code string }
	failures := []failure{}
	for _, item := range due {
		state, delay, code := "pending", "15 minutes", ""
		if requestErr != nil {
			delay, code = "5 minutes", "provider_unavailable"
			if item.attempts >= 2 {
				state = "failed"
			}
		} else if receipt, ok := payload.Data[item.ticket]; ok {
			switch receipt.Status {
			case "ok":
				state = "accepted"
			case "error":
				state = "rejected"
				code = strings.TrimSpace(receipt.Details.Error)
				if code == "" {
					code = "provider_error"
				}
				if len(code) > 64 {
					code = code[:64]
				}
			default:
				code = "unknown_receipt"
				if item.attempts >= 2 {
					state = "failed"
				}
			}
		} else {
			code = "receipt_missing"
			if item.attempts >= 2 {
				state = "missing"
			}
		}
		var next any
		if state == "pending" {
			next = delay
		}
		if _, err = tx.ExecContext(ctx, `UPDATE opportunity.notification_outbox SET receipt_status=$2,
 receipt_attempts=receipt_attempts+1,receipt_error=NULLIF($3,''),receipt_checked_at=now(),
 device_installation_id=CASE WHEN $2='pending' THEN device_installation_id ELSE NULL END,
 push_token_hash=CASE WHEN $2='pending' THEN push_token_hash ELSE NULL END,
 receipt_next_check_at=CASE WHEN $4::text IS NULL THEN NULL ELSE now()+$4::interval END,updated_at=now()
 WHERE offer_id=$1`, item.offerID, state, code, next); err != nil {
			return err
		}
		if state == "rejected" && code == "DeviceNotRegistered" && item.installationID != "" && item.tokenHash != "" {
			if _, err = tx.ExecContext(ctx, `UPDATE experience.devices SET push_token=NULL,notification_permission='denied',updated_at=now()
 WHERE installation_id=$1 AND user_id=$2 AND encode(digest(push_token,'sha256'),'hex')=$3`, item.installationID, item.userID, item.tokenHash); err != nil {
				return err
			}
		}
		if state == "rejected" || state == "missing" || state == "failed" {
			failures = append(failures, failure{item.offerID, state, code})
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	for _, item := range failures {
		slog.Warn("push receipt requires attention", "offer_id", item.offerID, "state", item.state, "code", item.code)
	}
	return requestErr
}
