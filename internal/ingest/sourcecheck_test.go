package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeCheckYAML(t *testing.T, lastChecked string) string {
	t.Helper()
	dir := t.TempDir()
	body := "last_checked: \"" + lastChecked + "\"\nmax_age_days: 45\nsources:\n" +
		"  - id: sample\n    url: https://www.gov.uk/guidance/example\n    must_contain:\n      - \"41,700\"\n"
	if err := os.WriteFile(filepath.Join(dir, "source_check.yaml"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckGovSources_markersPresent(t *testing.T) {
	dir := writeCheckYAML(t, time.Now().UTC().Format("2006-01-02"))
	err := checkGovSources(dir, func(string) (string, error) {
		return "general salary £41,700 per year", nil
	})
	if err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestCheckGovSources_markerMissing(t *testing.T) {
	dir := writeCheckYAML(t, time.Now().UTC().Format("2006-01-02"))
	err := checkGovSources(dir, func(string) (string, error) {
		return "general salary £38,700 per year", nil
	})
	if err == nil {
		t.Fatal("expected missing marker error")
	}
}

func TestCheckGovSources_staleLastChecked(t *testing.T) {
	dir := writeCheckYAML(t, "2025-01-01")
	err := checkGovSources(dir, func(string) (string, error) {
		return "41,700", nil
	})
	if err == nil {
		t.Fatal("expected stale last_checked error")
	}
}

func TestParseGovUKURL_rejectsOtherHosts(t *testing.T) {
	if _, err := parseGovUKURL("https://example.com/foo"); err == nil {
		t.Fatal("expected reject")
	}
	if _, err := parseGovUKURL("http://www.gov.uk/foo"); err == nil {
		t.Fatal("expected https only")
	}
	if _, err := parseGovUKURL("https://www.gov.uk/guidance/x"); err != nil {
		t.Fatalf("expected allow gov.uk, got %v", err)
	}
}

func TestCheckGovSources_fetchError(t *testing.T) {
	dir := writeCheckYAML(t, time.Now().UTC().Format("2006-01-02"))
	err := checkGovSources(dir, func(string) (string, error) {
		return "", errors.New("network down")
	})
	if err == nil {
		t.Fatal("expected fetch error")
	}
}
