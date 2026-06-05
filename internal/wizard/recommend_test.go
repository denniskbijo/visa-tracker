package wizard

import (
	"testing"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func TestRecommend_skilledWorkerWithSponsor(t *testing.T) {
	routes := []models.VisaRoute{{
		Slug: "skilled-worker", Name: "Skilled Worker",
		RequiresSponsor: true, SalaryThreshold: 3870000,
	}}
	recs := Recommend(Input{JobOffer: "yes", Sponsor: "yes", Salary: 45000}, routes)
	if len(recs) == 0 || recs[0].Route.Slug != "skilled-worker" {
		t.Fatalf("expected skilled-worker recommendation, got %+v", recs)
	}
	if recs[0].Match != MatchHigh {
		t.Fatalf("expected high match, got %s", recs[0].Match)
	}
}

func TestRecommend_ukGraduateNoOffer(t *testing.T) {
	routes := []models.VisaRoute{
		{Slug: "graduate", Name: "Graduate", DurationYears: 2},
		{Slug: "skilled-worker", Name: "Skilled Worker", RequiresSponsor: true},
	}
	recs := Recommend(Input{JobOffer: "no", Background: "uk_graduate"}, routes)
	if len(recs) == 0 || recs[0].Route.Slug != "graduate" {
		t.Fatalf("expected graduate first, got %+v", recs)
	}
}
