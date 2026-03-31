package models

import "time"

type ProcessingTime struct {
	ID           int64     `json:"id"`
	VisaRouteID  int64     `json:"visa_route_id"`
	MedianDays   int       `json:"median_days"`
	P90Days      int       `json:"p90_days"`
	SnapshotDate string    `json:"snapshot_date"`
	CreatedAt    time.Time `json:"created_at"`
}
