package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	if !strings.Contains(err.Error(), "days ago") {
		t.Fatalf("expected staleness error, got %v", err)
	}
}

func TestCheckGovSources_staleSkipsFetch(t *testing.T) {
	dir := writeCheckYAML(t, "2025-01-01")
	err := checkGovSources(dir, func(string) (string, error) {
		t.Fatal("fetch should not run when last_checked is stale")
		return "", nil
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

func TestCompareGovSources_matchWhenStale(t *testing.T) {
	dir := writeCheckYAML(t, "2025-01-01")
	report, err := compareGovSources(dir, func(string) (string, error) {
		return "general salary £41,700 per year", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Match() {
		t.Fatalf("expected match, missing %v", report.Missing)
	}
	if !report.Stale {
		t.Fatal("expected stale last_checked")
	}
	if report.AgeDays <= 45 {
		t.Fatalf("age_days = %d, want more than 45", report.AgeDays)
	}
	text := report.Format()
	if !strings.Contains(text, "status: match\n") || !strings.Contains(text, "stale: true\n") {
		t.Fatalf("report:\n%s", text)
	}
}

func TestCompareGovSources_mismatch(t *testing.T) {
	dir := writeCheckYAML(t, time.Now().UTC().Format("2006-01-02"))
	report, err := compareGovSources(dir, func(string) (string, error) {
		return "general salary £38,700 per year", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Match() {
		t.Fatal("expected mismatch")
	}
	if report.Stale {
		t.Fatal("fresh last_checked should not be stale")
	}
	if len(report.Missing) != 1 || !strings.Contains(report.Missing[0], "41,700") {
		t.Fatalf("missing = %#v", report.Missing)
	}
	text := report.Format()
	if !strings.Contains(text, "status: mismatch\n") || !strings.Contains(text, "missing:\n- ") {
		t.Fatalf("report:\n%s", text)
	}
}

func TestCompareGovSources_normalizesPoundEntities(t *testing.T) {
	dir := t.TempDir()
	body := "last_checked: \"" + time.Now().UTC().Format("2006-01-02") + "\"\nmax_age_days: 45\nsources:\n" +
		"  - id: sample\n    url: https://www.gov.uk/guidance/example\n    must_contain:\n      - \"£41,700\"\n"
	if err := os.WriteFile(filepath.Join(dir, "source_check.yaml"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := compareGovSources(dir, func(string) (string, error) {
		return "salary &#163;41,700 per year", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Match() {
		t.Fatalf("expected pound entity to match, missing %v", report.Missing)
	}
}

func TestCompareGovSources_staleAndMissing(t *testing.T) {
	dir := writeCheckYAML(t, "2025-01-01")
	report, err := compareGovSources(dir, func(string) (string, error) {
		return "no marker here", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Match() || !report.Stale {
		t.Fatalf("match=%v stale=%v missing=%v", report.Match(), report.Stale, report.Missing)
	}
}

func TestCompareGovSources_fetchError(t *testing.T) {
	dir := writeCheckYAML(t, time.Now().UTC().Format("2006-01-02"))
	_, err := compareGovSources(dir, func(string) (string, error) {
		return "", errors.New("network down")
	})
	if err == nil {
		t.Fatal("expected fetch error")
	}
	if !strings.Contains(err.Error(), "sample:") {
		t.Fatalf("expected source id in error, got %v", err)
	}
}
