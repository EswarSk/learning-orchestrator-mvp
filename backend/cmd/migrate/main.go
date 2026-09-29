package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"learning-orchestrator/backend/internal/platform"
)

func main() {
	db, err := platform.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	names, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(names, func(i, j int) bool { return names[i].Name() < names[j].Name() })
	ctx := context.Background()
	for _, entry := range names {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		data, err := os.ReadFile("migrations/" + entry.Name())
		if err != nil {
			log.Fatal(err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}
		if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(758923145)"); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		if _, err = tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS platform; CREATE TABLE IF NOT EXISTS platform.migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		var exists bool
		if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM platform.migrations WHERE name=$1)", entry.Name()).Scan(&exists); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		if !exists {
			if _, err = tx.ExecContext(ctx, string(data)); err != nil {
				tx.Rollback()
				log.Fatal(fmt.Errorf("%s: %w", entry.Name(), err))
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO platform.migrations(name) VALUES($1)", entry.Name()); err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
			log.Printf("applied %s", entry.Name())
		}
		if err = tx.Commit(); err != nil {
			log.Fatal(err)
		}
	}
}
