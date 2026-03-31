package store

import (
	"fmt"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (db *DB) ClearSponsors() error {
	_, err := db.conn.Exec(`DELETE FROM sponsors`)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(`DELETE FROM sponsors_fts`)
	return err
}

func (db *DB) InsertSponsor(s *models.Sponsor) error {
	_, err := db.conn.Exec(`
		INSERT INTO sponsors (name, city, county, route, rating, sub_rating)
		VALUES (?, ?, ?, ?, ?, ?)`,
		s.Name, s.City, s.County, s.Route, s.Rating, s.SubRating,
	)
	return err
}

func (db *DB) BulkInsertSponsors(sponsors []models.Sponsor) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO sponsors (name, city, county, route, rating, sub_rating) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, s := range sponsors {
		if _, err := stmt.Exec(s.Name, s.City, s.County, s.Route, s.Rating, s.SubRating); err != nil {
			return fmt.Errorf("insert sponsor %q: %w", s.Name, err)
		}
	}

	return tx.Commit()
}

func (db *DB) SearchSponsors(query, routeFilter, cityFilter string, limit, offset int) (*models.SponsorSearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	result := &models.SponsorSearchResult{
		Query:       query,
		RouteFilter: routeFilter,
		CityFilter:  cityFilter,
	}

	var where []string
	var args []interface{}

	if query != "" {
		where = append(where, "s.id IN (SELECT rowid FROM sponsors_fts WHERE sponsors_fts MATCH ?)")
		args = append(args, sanitizeFTS(query))
	}
	if routeFilter != "" {
		where = append(where, "s.route LIKE ?")
		args = append(args, "%"+routeFilter+"%")
	}
	if cityFilter != "" {
		where = append(where, "LOWER(s.city) = LOWER(?)")
		args = append(args, cityFilter)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sponsors s %s", whereClause)
	if err := db.conn.QueryRow(countQuery, args...).Scan(&result.Total); err != nil {
		return nil, fmt.Errorf("count sponsors: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT s.id, s.name, s.city, s.county, s.route, s.rating, s.sub_rating, s.ingested_at
		FROM sponsors s %s
		ORDER BY s.name
		LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := db.conn.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query sponsors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Sponsor
		if err := rows.Scan(&s.ID, &s.Name, &s.City, &s.County, &s.Route, &s.Rating, &s.SubRating, &s.IngestedAt); err != nil {
			return nil, fmt.Errorf("scan sponsor: %w", err)
		}
		result.Sponsors = append(result.Sponsors, s)
	}
	return result, rows.Err()
}

func (db *DB) SponsorCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sponsors").Scan(&count)
	return count, err
}

func (db *DB) DistinctSponsorCities(limit int) ([]string, error) {
	rows, err := db.conn.Query(`SELECT city, COUNT(*) as cnt FROM sponsors WHERE city != '' GROUP BY city ORDER BY cnt DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		var cnt int
		if err := rows.Scan(&city, &cnt); err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, rows.Err()
}

func sanitizeFTS(q string) string {
	q = strings.TrimSpace(q)
	words := strings.Fields(q)
	for i, w := range words {
		w = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, w)
		if w != "" {
			words[i] = w + "*"
		}
	}
	return strings.Join(words, " ")
}
