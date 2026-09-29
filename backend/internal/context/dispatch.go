package contextsignal

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"learning-orchestrator/backend/internal/contracts"
	"learning-orchestrator/backend/internal/platform"
)

func Dispatch(ctx context.Context, db *sql.DB, c client.Client, offers platform.Client) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := dispatchOne(ctx, db, c, offers); err != nil {
				slog.Error("context dispatch failed", "error", err)
			}
		}
	}
}

func dispatchOne(ctx context.Context, db *sql.DB, c client.Client, offers platform.Client) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var outboxID, eventID, userID, kind string
	var evaluateAt sql.NullTime
	var canceled bool
	var validUntil time.Time
	err = tx.QueryRowContext(ctx, `SELECT o.id,e.id,e.user_id,o.kind,e.evaluate_at,e.canceled_at IS NOT NULL,e.valid_until
 FROM context.outbox o JOIN context.events e ON e.id=o.event_id
 WHERE o.state='pending' AND o.next_attempt_at<=now() ORDER BY o.next_attempt_at LIMIT 1 FOR UPDATE OF o,e SKIP LOCKED`).Scan(&outboxID, &eventID, &userID, &kind, &evaluateAt, &canceled, &validUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if kind == "cancel" {
		err = offers.Do(ctx, userID, "POST", "/internal/offers/by-context/"+eventID+"/cancel", nil, nil)
		if err == nil {
			err = c.CancelWorkflow(ctx, "context-opportunity:"+eventID, "")
			var notFound *serviceerror.NotFound
			if errors.As(err, &notFound) {
				err = nil
			}
		}
	} else if !canceled && validUntil.After(time.Now()) {
		delay, _ := strconv.Atoi(platform.Env("OPPORTUNITY_DELAY_SECONDS", "120"))
		_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: "context-opportunity:" + eventID, TaskQueue: contracts.OpportunityTaskQueue}, contracts.OpportunityWorkflow, contracts.ContextTrigger{UserID: userID, EventID: eventID, DelaySeconds: delay, EvaluateAt: evaluateAt.Time})
	}
	var already *serviceerror.WorkflowExecutionAlreadyStarted
	if err != nil && !(kind == "observe" && errors.As(err, &already)) {
		_, _ = tx.ExecContext(ctx, "UPDATE context.outbox SET attempts=attempts+1,next_attempt_at=now()+interval '30 seconds' WHERE id=$1", outboxID)
		if commitErr := tx.Commit(); commitErr != nil {
			return commitErr
		}
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE context.outbox SET state='sent',attempts=attempts+1 WHERE id=$1", outboxID); err != nil {
		return err
	}
	return tx.Commit()
}
