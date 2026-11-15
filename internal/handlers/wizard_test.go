package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRedirectQuiz(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/wizard", nil)
	rr := httptest.NewRecorder()
	h.redirectQuiz(rr, req)
	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("status %d, want 301", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/quiz" {
		t.Fatalf("location %q, want /quiz", loc)
	}

	req = httptest.NewRequest(http.MethodGet, "/wizard?from=old", nil)
	rr = httptest.NewRecorder()
	h.redirectQuiz(rr, req)
	if loc := rr.Header().Get("Location"); loc != "/quiz?from=old" {
		t.Fatalf("location %q, want query preserved", loc)
	}

	req = httptest.NewRequest(http.MethodPost, "/wizard", nil)
	rr = httptest.NewRecorder()
	h.redirectQuiz(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /wizard status %d, want 405", rr.Code)
	}
}

func TestQuizInput(t *testing.T) {
	form := url.Values{
		"on_skilled_worker": {"yes"},
		"years_on_route":    {"5_plus"},
		"next_step":         {"global_talent"},
		"salary":            {"45000"},
	}
	req := httptest.NewRequest(http.MethodPost, "/quiz/results", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	in := quizInput(req)
	if in.OnSkilledWorker != "yes" || in.YearsOnRoute != "5_plus" || in.NextStep != "global_talent" || in.Salary != 45000 {
		t.Fatalf("parsed input %+v", in)
	}

	form.Set("salary", "-12")
	form.Set("on_skilled_worker", "not_yet")
	form.Set("offer_sponsor", "no_offer")
	form.Set("background", "uk_graduate")
	req = httptest.NewRequest(http.MethodPost, "/quiz/results", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	in = quizInput(req)
	if in.Salary != 0 || in.OfferSponsor != "no_offer" || in.Background != "uk_graduate" {
		t.Fatalf("expected clamped salary and not-yet fields, got %+v", in)
	}
}
