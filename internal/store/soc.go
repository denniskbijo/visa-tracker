package store

import (
	"database/sql"
	"fmt"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (db *DB) UpsertSOCCode(s *models.SOCCode) error {
	_, err := db.conn.Exec(`
		INSERT INTO soc_codes (code, title, description, going_rate_pence, on_immigration_salary_list)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(code) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			going_rate_pence = excluded.going_rate_pence,
			on_immigration_salary_list = excluded.on_immigration_salary_list,
			updated_at = CURRENT_TIMESTAMP`,
		s.Code, s.Title, s.Description, s.GoingRatePence, boolToInt(s.OnImmigrationSalaryList),
	)
	return err
}

func (db *DB) AllSOCCodes() ([]models.SOCCode, error) {
	rows, err := db.conn.Query(`SELECT code, title, description, going_rate_pence, on_immigration_salary_list, updated_at FROM soc_codes ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("query soc codes: %w", err)
	}
	defer rows.Close()

	var codes []models.SOCCode
	for rows.Next() {
		var s models.SOCCode
		var onList int
		if err := rows.Scan(&s.Code, &s.Title, &s.Description, &s.GoingRatePence, &onList, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan soc code: %w", err)
		}
		s.OnImmigrationSalaryList = onList == 1
		codes = append(codes, s)
	}
	return codes, rows.Err()
}

func (db *DB) SearchSOCCodes(query string) ([]models.SOCCode, error) {
	q := "%" + query + "%"
	rows, err := db.conn.Query(`
		SELECT code, title, description, going_rate_pence, on_immigration_salary_list, updated_at
		FROM soc_codes
		WHERE code LIKE ? OR LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)
		ORDER BY code`, q, q, q)
	if err != nil {
		return nil, fmt.Errorf("search soc codes: %w", err)
	}
	defer rows.Close()

	var codes []models.SOCCode
	for rows.Next() {
		var s models.SOCCode
		var onList int
		if err := rows.Scan(&s.Code, &s.Title, &s.Description, &s.GoingRatePence, &onList, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.OnImmigrationSalaryList = onList == 1
		codes = append(codes, s)
	}
	return codes, rows.Err()
}

func (db *DB) GetSOCCode(code string) (*models.SOCCode, error) {
	var s models.SOCCode
	var onList int
	err := db.conn.QueryRow(`SELECT code, title, description, going_rate_pence, on_immigration_salary_list, updated_at FROM soc_codes WHERE code = ?`, code).
		Scan(&s.Code, &s.Title, &s.Description, &s.GoingRatePence, &onList, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.OnImmigrationSalaryList = onList == 1
	return &s, nil
}
