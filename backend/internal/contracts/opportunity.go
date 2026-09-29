package contracts

import "time"

const OpportunityTaskQueue = "opportunity-v1"
const OpportunityWorkflow = "OpportunityWorkflow"

type ContextTrigger struct {
	UserID       string
	EventID      string
	DelaySeconds int
	EvaluateAt   time.Time
}
