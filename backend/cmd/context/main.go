package main

import (
	"context"
	contextsignal "learning-orchestrator/backend/internal/context"
	"learning-orchestrator/backend/internal/platform"
	"log"
)

func main() {
	db, err := platform.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	temporalClient, err := platform.DialTemporal()
	if err != nil {
		log.Fatal(err)
	}
	defer temporalClient.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go contextsignal.Dispatch(ctx, db, temporalClient, platform.NewClient("OPPORTUNITY_URL", "http://localhost:3003"))
	s := contextsignal.Server{DB: db, Experience: platform.NewClient("EXPERIENCE_URL", "http://localhost:3000")}
	if err := platform.Serve(platform.Env("LISTEN_ADDR", ":3002"), s.Handler()); err != nil {
		log.Fatal(err)
	}
}
