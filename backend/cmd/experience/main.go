package main

import (
	"context"
	"learning-orchestrator/backend/internal/experience"
	"learning-orchestrator/backend/internal/platform"
	"log"
)

func main() {
	db, err := platform.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	verifier, err := experience.NewVerifier(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	s := experience.Server{DB: db, Auth: verifier, Learning: platform.NewClient("LEARNING_URL", "http://localhost:3001"), Context: platform.NewClient("CONTEXT_URL", "http://localhost:3002"), Opportunity: platform.NewClient("OPPORTUNITY_URL", "http://localhost:3003")}
	if err := platform.Serve(platform.Env("LISTEN_ADDR", ":3000"), s.Handler()); err != nil {
		log.Fatal(err)
	}
}
