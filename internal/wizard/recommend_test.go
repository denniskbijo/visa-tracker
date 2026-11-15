package wizard

import (
	"strings"
	"testing"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func sampleRoutes() []models.VisaRoute {
	return []models.VisaRoute{
		{
			Slug: "skilled-worker", Name: "Skilled Worker", RequiresSponsor: true,
			SalaryThreshold: 4170000, DurationYears: 5,
			Description: "The main work visa route. Requires a licensed sponsor and a salary meeting the general threshold or the going rate, whichever is higher.",
		},
		{
			Slug: "global-talent", Name: "Global Talent", RequiresEndorsement: true, DurationYears: 5,
			Description: "Digital technology endorsement is assessed by Tech Nation. No sponsor or salary floor needed.",
		},
		{
			Slug: "high-potential-individual", Name: "High Potential Individual", DurationYears: 2,
			Description: "Two years, or three years with a PhD. Annual application cap. Cannot be extended.",
		},
		{
			Slug: "scale-up", Name: "Scale-up", RequiresSponsor: true,
			SalaryThreshold: 3910000, DurationYears: 2,
			Description: "Requires sponsor for the first 6 months, then free to change employers.",
		},
		{
			Slug: "intra-company-transfer", Name: "Senior or Specialist Worker (ICT)", RequiresSponsor: true,
			SalaryThreshold: 5250000, DurationYears: 5,
			Description: "For multinational companies transferring established employees to a UK branch.",
		},
		{
			Slug: "graduate", Name: "Graduate", DurationYears: 2,
			Description: "No sponsor needed. Currently 2 years (3 years with a PhD). Applications from 1 January 2027 are expected to be 18 months for most non-PhD graduates.",
		},
	}
}

func slugsOf(recs []Recommendation) []string {
	out := make([]string, len(recs))
	for i, rec := range recs {
		out[i] = rec.Route.Slug
	}
	return out
}

func bySlug(recs []Recommendation, slug string) (Recommendation, bool) {
	for _, rec := range recs {
		if rec.Route.Slug == slug {
			return rec, true
		}
	}
	return Recommendation{}, false
}

func joinedReasons(rec Recommendation) string {
	return strings.Join(rec.Reasons, " ")
}

func assertNoEmDash(t *testing.T, recs []Recommendation) {
	t.Helper()
	for _, rec := range recs {
		for _, reason := range rec.Reasons {
			if strings.Contains(reason, "\u2014") || strings.Contains(reason, " -- ") {
				t.Fatalf("reason contains an em dash: %s", reason)
			}
		}
	}
}

func TestRecommend_currentSkilledWorkerStay(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "yes",
		YearsOnRoute:    "under_3",
		NextStep:        "stay",
		Salary:          45000,
	}, sampleRoutes())
	assertNoEmDash(t, recs)
	if len(recs) != 1 || recs[0].Route.Slug != "skilled-worker" || recs[0].Match != MatchHigh {
		t.Fatalf("expected only skilled-worker high, got %+v", slugsOf(recs))
	}
	text := joinedReasons(recs[0])
	if !strings.Contains(text, "early on Skilled Worker") || !strings.Contains(text, "5 years") {
		t.Fatalf("expected ILR timeline reason, got %s", text)
	}
	if !strings.Contains(text, "meets the £41,700") {
		t.Fatalf("expected salary threshold from route data, got %s", text)
	}
	if strings.Contains(text, "intra-company") {
		t.Fatalf("ICT should not appear in skilled worker reasons: %s", text)
	}
}

func TestRecommend_currentSkilledWorkerILRReadyAndBelowSalary(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "yes",
		YearsOnRoute:    "5_plus",
		NextStep:        "stay",
		Salary:          30000,
	}, sampleRoutes())
	rec, ok := bySlug(recs, "skilled-worker")
	if !ok || rec.Match != MatchHigh {
		t.Fatalf("someone already on Skilled Worker stays a high match, got %+v", recs)
	}
	text := joinedReasons(rec)
	if !strings.Contains(text, "may already be at the 5-year mark") {
		t.Fatalf("expected 5-year readiness, got %s", text)
	}
	if !strings.Contains(text, "below the £41,700") {
		t.Fatalf("expected below-threshold salary check, got %s", text)
	}
}

func TestRecommend_currentSkilledWorkerMidILR(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "yes",
		YearsOnRoute:    "3_to_5",
		NextStep:        "stay",
	}, sampleRoutes())
	rec, ok := bySlug(recs, "skilled-worker")
	if !ok {
		t.Fatal("missing skilled worker")
	}
	text := joinedReasons(rec)
	if !strings.Contains(text, "later part of the 5-year") {
		t.Fatalf("expected mid-route ILR reason, got %s", text)
	}
	if !strings.Contains(text, "Confirm your salary meets the £41,700") {
		t.Fatalf("blank salary should ask them to confirm the stored threshold, got %s", text)
	}
}

func TestRecommend_currentSkilledWorkerSwitchGlobalTalent(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "yes",
		YearsOnRoute:    "3_to_5",
		NextStep:        "global_talent",
		Salary:          80000,
	}, sampleRoutes())
	if recs[0].Route.Slug != "skilled-worker" {
		t.Fatalf("skilled worker should stay first, got %v", slugsOf(recs))
	}
	gt, ok := bySlug(recs, "global-talent")
	if !ok || gt.Match != MatchMedium {
		t.Fatalf("expected global talent medium, got %+v", recs)
	}
	if !strings.Contains(joinedReasons(gt), "no salary floor") {
		t.Fatalf("expected no salary floor from route data, got %s", joinedReasons(gt))
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("ICT must not be offered as a switch from Skilled Worker")
	}
	if _, ok := bySlug(recs, "graduate"); ok {
		t.Fatal("graduate is not a switch for someone already on Skilled Worker")
	}
}

func TestRecommend_currentSkilledWorkerSwitchScaleUp(t *testing.T) {
	meet := Recommend(Input{
		OnSkilledWorker: "yes", YearsOnRoute: "under_3", NextStep: "scale_up", Salary: 40000,
	}, sampleRoutes())
	su, ok := bySlug(meet, "scale-up")
	if !ok || su.Match != MatchMedium {
		t.Fatalf("salary above scale-up threshold should be medium, got %+v", meet)
	}
	if !strings.Contains(joinedReasons(su), "£39,100") {
		t.Fatalf("expected scale-up threshold from data, got %s", joinedReasons(su))
	}

	below := Recommend(Input{
		OnSkilledWorker: "yes", YearsOnRoute: "under_3", NextStep: "scale_up", Salary: 30000,
	}, sampleRoutes())
	su, ok = bySlug(below, "scale-up")
	if !ok || su.Match != MatchLow {
		t.Fatalf("salary below scale-up threshold should be low, got %+v", below)
	}
	if _, ok := bySlug(below, "intra-company-transfer"); ok {
		t.Fatal("ICT must not appear on the Skilled Worker path")
	}
}

func TestRecommend_notYetGraduateNoOffer(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "no_offer",
		Background:      "uk_graduate",
	}, sampleRoutes())
	assertNoEmDash(t, recs)
	if len(recs) == 0 || recs[0].Route.Slug != "graduate" || recs[0].Match != MatchHigh {
		t.Fatalf("expected graduate first, got %+v", slugsOf(recs))
	}
	text := joinedReasons(recs[0])
	if !strings.Contains(text, "1 January 2027") || !strings.Contains(text, "PhD") {
		t.Fatalf("expected graduate seed wording, got %s", text)
	}
	if _, ok := bySlug(recs, "skilled-worker"); !ok {
		t.Fatal("expected a later skilled worker note")
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("ICT does not belong on the no-offer path")
	}
}

func TestRecommend_notYetTopUniversity(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "no_offer",
		Background:      "top_university",
	}, sampleRoutes())
	if recs[0].Route.Slug != "high-potential-individual" || recs[0].Match != MatchHigh {
		t.Fatalf("expected HPI first, got %+v", slugsOf(recs))
	}
	if !strings.Contains(joinedReasons(recs[0]), "three years with a PhD") {
		t.Fatalf("expected HPI seed wording, got %s", joinedReasons(recs[0]))
	}
	if _, ok := bySlug(recs, "global-talent"); !ok {
		t.Fatal("expected global talent alongside HPI")
	}
}

func TestRecommend_notYetTechTalent(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "no_offer",
		Background:      "tech_talent",
		Salary:          100000,
	}, sampleRoutes())
	if recs[0].Route.Slug != "global-talent" || recs[0].Match != MatchHigh {
		t.Fatalf("expected global talent first, got %+v", slugsOf(recs))
	}
	if _, ok := bySlug(recs, "high-potential-individual"); !ok {
		t.Fatal("expected HPI as a secondary check")
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("a high salary without a sponsor is not an ICT case")
	}
}

func TestRecommend_notYetNoOfferOther(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "no_offer",
		Background:      "none",
	}, sampleRoutes())
	got := map[string]bool{}
	for _, rec := range recs {
		got[rec.Route.Slug] = true
	}
	for _, slug := range []string{"global-talent", "high-potential-individual", "graduate"} {
		if !got[slug] {
			t.Fatalf("expected %s among %v", slug, slugsOf(recs))
		}
	}
	if got["intra-company-transfer"] || got["scale-up"] {
		t.Fatalf("sponsored routes should not lead the no-offer path, got %v", slugsOf(recs))
	}
}

func TestRecommend_notYetOfferAndSponsor(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_yes",
		Salary:          45000,
		Background:      "uk_graduate",
	}, sampleRoutes())
	sw, ok := bySlug(recs, "skilled-worker")
	if !ok || sw.Match != MatchHigh {
		t.Fatalf("expected skilled worker high, got %+v", recs)
	}
	if !strings.Contains(joinedReasons(sw), "meets the £41,700") {
		t.Fatalf("expected threshold check, got %s", joinedReasons(sw))
	}
	su, ok := bySlug(recs, "scale-up")
	if !ok || su.Match != MatchMedium {
		t.Fatalf("expected scale-up medium, got %+v", recs)
	}
	if _, ok := bySlug(recs, "graduate"); ok {
		t.Fatal("a sponsored offer should not also push Graduate")
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("£45,000 is below the ICT threshold, so ICT must not appear")
	}
}

func TestRecommend_notYetOfferBelowThresholds(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_yes",
		Salary:          20000,
	}, sampleRoutes())
	sw, ok := bySlug(recs, "skilled-worker")
	if !ok || sw.Match != MatchLow {
		t.Fatalf("expected skilled worker low, got %+v", recs)
	}
	su, ok := bySlug(recs, "scale-up")
	if !ok || su.Match != MatchLow {
		t.Fatalf("expected scale-up low, got %+v", recs)
	}
	if !strings.Contains(joinedReasons(sw), "below the £41,700") {
		t.Fatalf("expected skilled worker shortfall, got %s", joinedReasons(sw))
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("ICT must not appear below its threshold")
	}
}

func TestRecommend_ictStatesItDoesNotLeadToILR(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_yes",
		Salary:          60000,
		Background:      "tech_talent",
	}, sampleRoutes())
	ict, ok := bySlug(recs, "intra-company-transfer")
	if !ok || ict.Match != MatchLow {
		t.Fatalf("expected ICT as a low match when salary meets £52,500, got %+v", recs)
	}
	text := strings.ToLower(joinedReasons(ict))
	if !strings.Contains(text, "does not lead to indefinite leave to remain (ilr)") {
		t.Fatalf("ICT must say it does not lead to ILR, got %s", joinedReasons(ict))
	}
	if !strings.Contains(text, "does not count toward settlement") {
		t.Fatalf("ICT must say time does not count toward settlement, got %s", joinedReasons(ict))
	}
	if strings.Contains(text, "route to ilr") || strings.Contains(text, "path to ilr") || strings.Contains(text, "leads to ilr") {
		t.Fatalf("ICT wording still presents it as an ILR route: %s", joinedReasons(ict))
	}
	sw, ok := bySlug(recs, "skilled-worker")
	if !ok || sw.Match != MatchHigh || matchRank(ict.Match) >= matchRank(sw.Match) {
		t.Fatalf("ICT must not outrank skilled worker, got %+v", slugsOf(recs))
	}
	if !strings.Contains(joinedReasons(ict), "£52,500") {
		t.Fatalf("expected ICT threshold from data, got %s", joinedReasons(ict))
	}
}

func TestRecommend_notYetUnsureSponsor(t *testing.T) {
	recs := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_unsure",
		Salary:          45000,
	}, sampleRoutes())
	sw, ok := bySlug(recs, "skilled-worker")
	if !ok || sw.Match != MatchMedium {
		t.Fatalf("expected skilled worker medium, got %+v", recs)
	}
	if !strings.Contains(joinedReasons(sw), "Register of Licensed Sponsors") {
		t.Fatalf("expected sponsor-check reason, got %s", joinedReasons(sw))
	}
	if _, ok := bySlug(recs, "intra-company-transfer"); ok {
		t.Fatal("do not suggest ICT until the sponsor is known")
	}
}

func TestRecommend_notYetOfferNoSponsor(t *testing.T) {
	grad := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_no",
		Background:      "uk_graduate",
		Salary:          50000,
	}, sampleRoutes())
	if grad[0].Route.Slug != "graduate" {
		t.Fatalf("UK graduate without a sponsor should see Graduate first, got %v", slugsOf(grad))
	}
	if _, ok := bySlug(grad, "intra-company-transfer"); ok {
		t.Fatal("no sponsor means no ICT")
	}

	talent := Recommend(Input{
		OnSkilledWorker: "not_yet",
		OfferSponsor:    "sponsor_no",
		Background:      "tech_talent",
	}, sampleRoutes())
	if talent[0].Route.Slug != "global-talent" || talent[0].Match != MatchHigh {
		t.Fatalf("expected global talent high, got %+v", slugsOf(talent))
	}
	sw, ok := bySlug(talent, "skilled-worker")
	if !ok || sw.Match != MatchLow {
		t.Fatalf("expected skilled worker low without a sponsor, got %+v", talent)
	}
}

func TestRecommend_skilledWorkerPathNeverIncludesICT(t *testing.T) {
	steps := []string{"stay", "global_talent", "scale_up"}
	years := []string{"under_3", "3_to_5", "5_plus", ""}
	salaries := []int64{0, 20000, 45000, 60000, 120000}
	for _, step := range steps {
		for _, year := range years {
			for _, salary := range salaries {
				recs := Recommend(Input{
					OnSkilledWorker: "yes",
					YearsOnRoute:    year,
					NextStep:        step,
					Salary:          salary,
				}, sampleRoutes())
				if _, ok := bySlug(recs, "intra-company-transfer"); ok {
					t.Fatalf("ICT on skilled worker path step=%s years=%s salary=%d", step, year, salary)
				}
				if len(recs) == 0 || recs[0].Route.Slug != "skilled-worker" {
					t.Fatalf("skilled worker should lead, got %v", slugsOf(recs))
				}
			}
		}
	}
}
