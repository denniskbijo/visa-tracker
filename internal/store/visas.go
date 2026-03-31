package store

import (
	"database/sql"
	"fmt"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (db *DB) UpsertVisaRoute(v *models.VisaRoute) error {
	_, err := db.conn.Exec(`
		INSERT INTO visa_routes (slug, name, description, requires_sponsor, requires_endorsement, salary_threshold, duration_years)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			requires_sponsor = excluded.requires_sponsor,
			requires_endorsement = excluded.requires_endorsement,
			salary_threshold = excluded.salary_threshold,
			duration_years = excluded.duration_years,
			updated_at = CURRENT_TIMESTAMP`,
		v.Slug, v.Name, v.Description,
		boolToInt(v.RequiresSponsor), boolToInt(v.RequiresEndorsement),
		v.SalaryThreshold, v.DurationYears,
	)
	return err
}

func (db *DB) AllVisaRoutes() ([]models.VisaRoute, error) {
	rows, err := db.conn.Query(`SELECT id, slug, name, description, requires_sponsor, requires_endorsement, salary_threshold, duration_years, updated_at FROM visa_routes ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query visa routes: %w", err)
	}
	defer rows.Close()

	var routes []models.VisaRoute
	for rows.Next() {
		var v models.VisaRoute
		var sponsor, endorsement int
		if err := rows.Scan(&v.ID, &v.Slug, &v.Name, &v.Description, &sponsor, &endorsement, &v.SalaryThreshold, &v.DurationYears, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan visa route: %w", err)
		}
		v.RequiresSponsor = sponsor == 1
		v.RequiresEndorsement = endorsement == 1
		routes = append(routes, v)
	}
	return routes, rows.Err()
}

func (db *DB) GetVisaRouteBySlug(slug string) (*models.VisaRoute, error) {
	var v models.VisaRoute
	var sponsor, endorsement int
	err := db.conn.QueryRow(`SELECT id, slug, name, description, requires_sponsor, requires_endorsement, salary_threshold, duration_years, updated_at FROM visa_routes WHERE slug = ?`, slug).
		Scan(&v.ID, &v.Slug, &v.Name, &v.Description, &sponsor, &endorsement, &v.SalaryThreshold, &v.DurationYears, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get visa route: %w", err)
	}
	v.RequiresSponsor = sponsor == 1
	v.RequiresEndorsement = endorsement == 1
	return &v, nil
}

func (db *DB) InsertThreshold(t *models.SalaryThreshold) error {
	_, err := db.conn.Exec(`
		INSERT INTO salary_thresholds (visa_route_id, soc_code, amount_pence, effective_date, notes)
		VALUES (?, ?, ?, ?, ?)`,
		t.VisaRouteID, t.SOCCode, t.AmountPence, t.EffectiveDate, t.Notes,
	)
	return err
}

func (db *DB) LatestThresholds(visaRouteID int64) ([]models.SalaryThreshold, error) {
	rows, err := db.conn.Query(`
		SELECT id, visa_route_id, soc_code, amount_pence, effective_date, notes, created_at
		FROM salary_thresholds
		WHERE visa_route_id = ?
		ORDER BY effective_date DESC
		LIMIT 10`, visaRouteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var thresholds []models.SalaryThreshold
	for rows.Next() {
		var t models.SalaryThreshold
		if err := rows.Scan(&t.ID, &t.VisaRouteID, &t.SOCCode, &t.AmountPence, &t.EffectiveDate, &t.Notes, &t.CreatedAt); err != nil {
			return nil, err
		}
		thresholds = append(thresholds, t)
	}
	return thresholds, rows.Err()
}

func (db *DB) InsertProcessingTime(pt *models.ProcessingTime) error {
	_, err := db.conn.Exec(`
		INSERT INTO processing_times (visa_route_id, median_days, p90_days, snapshot_date)
		VALUES (?, ?, ?, ?)`,
		pt.VisaRouteID, pt.MedianDays, pt.P90Days, pt.SnapshotDate,
	)
	return err
}

func (db *DB) LatestProcessingTime(visaRouteID int64) (*models.ProcessingTime, error) {
	var pt models.ProcessingTime
	err := db.conn.QueryRow(`
		SELECT id, visa_route_id, median_days, p90_days, snapshot_date, created_at
		FROM processing_times
		WHERE visa_route_id = ?
		ORDER BY snapshot_date DESC
		LIMIT 1`, visaRouteID).
		Scan(&pt.ID, &pt.VisaRouteID, &pt.MedianDays, &pt.P90Days, &pt.SnapshotDate, &pt.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
