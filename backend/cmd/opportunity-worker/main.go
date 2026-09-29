package main

import (
	"context"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"learning-orchestrator/backend/internal/contracts"
	"learning-orchestrator/backend/internal/opportunity"
	"learning-orchestrator/backend/internal/platform"
	"log"
)

func main() {
	db, err := platform.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	c, err := platform.DialTemporal()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	w := worker.New(c, contracts.OpportunityTaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(opportunity.Workflow, workflow.RegisterOptions{Name: contracts.OpportunityWorkflow})
	a := opportunity.Activities{DB: db, Context: platform.NewClient("CONTEXT_URL", "http://localhost:3002"), Experience: platform.NewClient("EXPERIENCE_URL", "http://localhost:3000"), Learning: platform.NewClient("LEARNING_URL", "http://localhost:3001")}
	w.RegisterActivityWithOptions(a.Evaluate, activity.RegisterOptions{Name: "EvaluateOpportunity"})
	if platform.Env("PUSH_ENABLED", "false") == "true" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go opportunity.DispatchNotifications(ctx, db)
	}
	log.Fatal(w.Run(worker.InterruptCh()))
}
