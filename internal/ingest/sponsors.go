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

const maxSponsorCSVSize = 50 * 1024 * 1024 // 50 MB sanity limit
const insertBatchSize = 2000

func (ing *Ingester) RefreshSponsors() error {
	url := ing.cfg.SponsorCSVURL
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("sponsor CSV URL must use HTTPS, got: %s", url)
	}

	log.Printf("downloading sponsor list from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download sponsors CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from sponsor CSV", resp.StatusCode)
	}

	limitedBody := io.LimitReader(resp.Body, maxSponsorCSVSize)
	total, err := ing.streamSponsorsCSV(limitedBody)
	if err != nil {
		return fmt.Errorf("stream sponsors CSV: %w", err)
	}

	log.Printf("loaded %d sponsors into database", total)
	return nil
}

// streamSponsorsCSV reads the CSV row-by-row, inserting in batches to
// keep memory usage constant regardless of file size.
func (ing *Ingester) streamSponsorsCSV(r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
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

	if err := ing.db.ClearSponsors(); err != nil {
		return 0, fmt.Errorf("clear sponsors: %w", err)
	}

	batch := make([]models.Sponsor, 0, insertBatchSize)
	total := 0

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

		batch = append(batch, s)
		if len(batch) >= insertBatchSize {
			if err := ing.db.BulkInsertSponsors(batch); err != nil {
				return total, fmt.Errorf("bulk insert at row %d: %w", total, err)
			}
			total += len(batch)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := ing.db.BulkInsertSponsors(batch); err != nil {
			return total, fmt.Errorf("bulk insert final batch: %w", err)
		}
		total += len(batch)
	}

	if total < 1000 {
		return total, fmt.Errorf("suspiciously few sponsors (%d), data may be corrupted", total)
	}

	return total, nil
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
