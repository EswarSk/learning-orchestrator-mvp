package main

import (
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
	if err := platform.Serve(platform.Env("LISTEN_ADDR", ":3003"), opportunity.Server{DB: db}.Handler()); err != nil {
		log.Fatal(err)
	}
}
