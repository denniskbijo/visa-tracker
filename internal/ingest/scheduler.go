package ingest

import (
	"log"
	"time"
)

func (ing *Ingester) StartScheduler() {
	interval := time.Duration(ing.cfg.RefreshInterval) * time.Hour
	log.Printf("sponsor refresh scheduler started (every %s)", interval)

	count, _ := ing.db.SponsorCount()
	if count == 0 {
		log.Println("no sponsors in DB, running initial ingestion")
		if err := ing.RefreshSponsors(); err != nil {
			log.Printf("initial sponsor ingestion failed: %v", err)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("scheduled sponsor refresh starting")
		if err := ing.RefreshSponsors(); err != nil {
			log.Printf("scheduled sponsor refresh failed: %v", err)
		}
	}
}
