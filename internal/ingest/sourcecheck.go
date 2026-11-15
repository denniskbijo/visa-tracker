package ingest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/denniskbijo/visa-tracker/internal/fsutil"
	"gopkg.in/yaml.v3"
)

type SourceCheckFile struct {
	LastChecked string             `yaml:"last_checked"`
	MaxAgeDays  int                `yaml:"max_age_days"`
	Sources     []SourceCheckEntry `yaml:"sources"`
}

type SourceCheckEntry struct {
	ID          string   `yaml:"id"`
	URL         string   `yaml:"url"`
	MustContain []string `yaml:"must_contain"`
}

// SourceReport is the marker comparison for one run.
// Stale is separate from a mismatch: a stale last_checked does not mean the pages differ.
type SourceReport struct {
	LastChecked string
	MaxAgeDays  int
	AgeDays     int
	Stale       bool
	Missing     []string
}

// Match reports whether every expected marker was found.
// It ignores staleness.
func (r SourceReport) Match() bool {
	return len(r.Missing) == 0
}

// Format is the line-oriented report printed by -report-sources.
func (r SourceReport) Format() string {
	status := "match"
	if !r.Match() {
		status = "mismatch"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "status: %s\n", status)
	fmt.Fprintf(&b, "last_checked: %s\n", r.LastChecked)
	fmt.Fprintf(&b, "age_days: %d\n", r.AgeDays)
	fmt.Fprintf(&b, "max_age_days: %d\n", r.MaxAgeDays)
	fmt.Fprintf(&b, "stale: %t\n", r.Stale)
	if len(r.Missing) > 0 {
		b.WriteString("missing:\n")
		for _, line := range r.Missing {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	}
	return b.String()
}

func (f SourceCheckFile) LastCheckedDisplay() string {
	t, err := time.Parse("2006-01-02", f.LastChecked)
	if err != nil {
		return f.LastChecked
	}
	return t.Format("January 2006")
}

func LoadSourceCheck(dataDir string) (*SourceCheckFile, error) {
	data, err := fsutil.ReadFileUnderRoot(dataDir, "source_check.yaml")
	if err != nil {
		return nil, fmt.Errorf("read source_check.yaml: %w", err)
	}
	var f SourceCheckFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse source_check.yaml: %w", err)
	}
	if f.MaxAgeDays <= 0 {
		f.MaxAgeDays = 45
	}
	return &f, nil
}

// CheckGovSources fetches the official appendices and fails if expected
// figures are missing or last_checked is older than max_age_days.
// Staleness is checked before any page is fetched.
func CheckGovSources(dataDir string) error {
	return checkGovSources(dataDir, fetchGovUK)
}

func checkGovSources(dataDir string, fetch func(rawURL string) (string, error)) error {
	f, err := LoadSourceCheck(dataDir)
	if err != nil {
		return err
	}
	age, err := lastCheckedAgeDays(f.LastChecked)
	if err != nil {
		return err
	}
	if age > f.MaxAgeDays {
		return fmt.Errorf("last_checked %s is %d days ago (limit %d); re-read gov.uk and bump last_checked in data/source_check.yaml",
			f.LastChecked, age, f.MaxAgeDays)
	}
	missing, err := missingMarkers(f, fetch)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("gov.uk pages no longer match our snapshot; update data/ YAML and source_check.yaml:\n  %s",
			strings.Join(missing, "\n  "))
	}
	return nil
}

// CompareGovSources fetches the official appendices and reports whether the
// expected markers are still present. A stale last_checked is included in
// the report and is not an error.
func CompareGovSources(dataDir string) (*SourceReport, error) {
	return compareGovSources(dataDir, fetchGovUK)
}

func compareGovSources(dataDir string, fetch func(rawURL string) (string, error)) (*SourceReport, error) {
	f, err := LoadSourceCheck(dataDir)
	if err != nil {
		return nil, err
	}
	age, err := lastCheckedAgeDays(f.LastChecked)
	if err != nil {
		return nil, err
	}
	missing, err := missingMarkers(f, fetch)
	if err != nil {
		return nil, err
	}
	return &SourceReport{
		LastChecked: f.LastChecked,
		MaxAgeDays:  f.MaxAgeDays,
		AgeDays:     age,
		Stale:       age > f.MaxAgeDays,
		Missing:     missing,
	}, nil
}

func lastCheckedAgeDays(lastChecked string) (int, error) {
	checked, err := time.Parse("2006-01-02", lastChecked)
	if err != nil {
		return 0, fmt.Errorf("last_checked %q: %w", lastChecked, err)
	}
	return int(time.Since(checked).Hours() / 24), nil
}

func missingMarkers(f *SourceCheckFile, fetch func(rawURL string) (string, error)) ([]string, error) {
	var missing []string
	for _, src := range f.Sources {
		body, err := fetch(src.URL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.ID, err)
		}
		body = normalizeGovHTML(body)
		for _, marker := range src.MustContain {
			if !strings.Contains(body, marker) {
				missing = append(missing, fmt.Sprintf("%s: missing %q", src.ID, marker))
			}
		}
	}
	return missing, nil
}

func normalizeGovHTML(s string) string {
	s = strings.ReplaceAll(s, "&pound;", "£")
	s = strings.ReplaceAll(s, "&#163;", "£")
	return s
}

func fetchGovUK(raw string) (string, error) {
	u, err := parseGovUKURL(raw)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req.URL.Scheme != "https" || !isGovUKHost(req.URL.Hostname()) {
				return fmt.Errorf("redirect host not allowed: %s", req.URL.Host)
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "visa-tracker-source-check (+https://github.com/denniskbijo/visa-tracker)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", u.Host+u.Path, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func parseGovUKURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("URL must use HTTPS")
	}
	if !isGovUKHost(u.Hostname()) {
		return nil, fmt.Errorf("host must be www.gov.uk, got %q", u.Hostname())
	}
	if u.Path == "" || u.Path == "/" {
		return nil, fmt.Errorf("URL must include a path")
	}
	return u, nil
}

func isGovUKHost(host string) bool {
	return strings.ToLower(host) == "www.gov.uk"
}
