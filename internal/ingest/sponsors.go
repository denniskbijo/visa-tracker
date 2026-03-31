package ingest

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (ing *Ingester) RefreshSponsors() error {
	log.Printf("downloading sponsor list from %s", ing.cfg.SponsorCSVURL)

	resp, err := http.Get(ing.cfg.SponsorCSVURL)
	if err != nil {
		return fmt.Errorf("download sponsors CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from sponsor CSV", resp.StatusCode)
	}

	sponsors, err := parseSponsorsCSV(resp.Body)
	if err != nil {
		return fmt.Errorf("parse sponsors CSV: %w", err)
	}

	log.Printf("parsed %d sponsors, loading into database", len(sponsors))

	if err := ing.db.ClearSponsors(); err != nil {
		return fmt.Errorf("clear sponsors: %w", err)
	}

	const batchSize = 5000
	for i := 0; i < len(sponsors); i += batchSize {
		end := i + batchSize
		if end > len(sponsors) {
			end = len(sponsors)
		}
		if err := ing.db.BulkInsertSponsors(sponsors[i:end]); err != nil {
			return fmt.Errorf("bulk insert sponsors batch %d: %w", i/batchSize, err)
		}
	}

	log.Printf("loaded %d sponsors into database", len(sponsors))
	return nil
}

func parseSponsorsCSV(r io.Reader) ([]models.Sponsor, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	nameIdx := findCol(colIdx, "organisation name", "name")
	cityIdx := findCol(colIdx, "city / town", "town/city", "city", "town")
	countyIdx := findCol(colIdx, "county", "region")
	routeIdx := findCol(colIdx, "route", "type & rating", "type")
	ratingIdx := findCol(colIdx, "rating", "type & rating")
	subIdx := findCol(colIdx, "sub rating")

	var sponsors []models.Sponsor
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		s := models.Sponsor{
			Name:      getField(record, nameIdx),
			City:      getField(record, cityIdx),
			County:    getField(record, countyIdx),
			Route:     getField(record, routeIdx),
			Rating:    getField(record, ratingIdx),
			SubRating: getField(record, subIdx),
		}

		if s.Name == "" {
			continue
		}
		sponsors = append(sponsors, s)
	}

	return sponsors, nil
}

func findCol(idx map[string]int, names ...string) int {
	for _, n := range names {
		if i, ok := idx[n]; ok {
			return i
		}
	}
	return -1
}

func getField(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}
