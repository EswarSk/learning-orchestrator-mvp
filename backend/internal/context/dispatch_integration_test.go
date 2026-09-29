package contextsignal

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"learning-orchestrator/backend/internal/contracts"
	"learning-orchestrator/backend/internal/opportunity"
	"learning-orchestrator/backend/internal/platform"
)

func TestDispatchCancelsWaitingTemporalWorkflow(t *testing.T) {
	dsn, address := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_TEMPORAL_ADDRESS")
	if dsn == "" || address == "" {
		t.Skip("set TEST_DATABASE_URL and TEST_TEMPORAL_ADDRESS for local Temporal integration")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a disposable database ending in _test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	w := worker.New(c, contracts.OpportunityTaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(opportunity.Workflow, workflow.RegisterOptions{Name: contracts.OpportunityWorkflow})
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	var cancels atomic.Int32
	offers := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || !strings.HasSuffix(request.URL.Path, "/cancel") {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		cancels.Add(1)
		response.WriteHeader(http.StatusOK)
	}))
	defer offers.Close()
	user := "temporal-dispatch-test-" + time.Now().UTC().Format("20060102150405.000000000")
	var eventID string
	err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
	 VALUES($1,'calendar',$1,'test-calendar','practice_place',now(),now()+interval '1 hour',now()+interval '2 hours') RETURNING id`, user).Scan(&eventID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = c.CancelWorkflow(context.Background(), "context-opportunity:"+eventID, "")
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id=$1", eventID)
		_, _ = db.Exec("DELETE FROM context.events WHERE id=$1", eventID)
	}()
	if _, err := db.Exec(`INSERT INTO context.outbox(event_id,kind,next_attempt_at) VALUES($1,'observe','2000-01-01')`, eventID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	provider := platform.Client{Base: offers.URL, HTTP: offers.Client()}
	if err := dispatchOne(ctx, db, c, provider); err != nil {
		t.Fatal(err)
	}
	workflowID := "context-opportunity:" + eventID
	for ctx.Err() == nil {
		started, err := c.DescribeWorkflowExecution(ctx, workflowID, "")
		if err != nil {
			t.Fatal(err)
		}
		if started.WorkflowExecutionInfo.Status != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
			t.Fatalf("waiting workflow not running: %v", started.WorkflowExecutionInfo.Status)
		}
		history := c.GetWorkflowHistory(ctx, workflowID, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		waiting := false
		for history.HasNext() {
			event, err := history.Next()
			if err != nil {
				t.Fatal(err)
			}
			waiting = waiting || event.EventType == enumspb.EVENT_TYPE_TIMER_STARTED
		}
		if waiting {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if ctx.Err() != nil {
		t.Fatal("workflow did not begin its scheduled wait")
	}
	if _, err := db.Exec(`UPDATE context.events SET canceled_at=now() WHERE id=$1`, eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO context.outbox(event_id,kind,next_attempt_at) VALUES($1,'cancel','2000-01-01')`, eventID); err != nil {
		t.Fatal(err)
	}
	if err := dispatchOne(ctx, db, c, provider); err != nil {
		t.Fatal(err)
	}
	for ctx.Err() == nil {
		state, err := c.DescribeWorkflowExecution(ctx, workflowID, "")
		if err != nil {
			t.Fatal(err)
		}
		if state.WorkflowExecutionInfo.Status == enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED {
			if cancels.Load() != 1 {
				t.Fatalf("offer cancellation calls = %d, want 1", cancels.Load())
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("waiting workflow did not cancel before timeout")
}
