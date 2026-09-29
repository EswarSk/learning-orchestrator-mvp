package main

import (
	"learning-orchestrator/backend/internal/learning"
	"learning-orchestrator/backend/internal/platform"
	"log"
)

func main() {
	db, err := platform.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	catalog, err := learning.LoadCatalog(platform.Env("CONTENT_DIR", "../content"))
	if err != nil {
		log.Fatal(err)
	}
	if err := platform.Serve(platform.Env("LISTEN_ADDR", ":3001"), learning.Server{DB: db, Catalog: catalog, AI: learning.AIFromEnv()}.Handler()); err != nil {
		log.Fatal(err)
	}
}
