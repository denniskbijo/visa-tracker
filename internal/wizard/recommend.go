package wizard

import (
	"sort"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

type Input struct {
	JobOffer   string // yes, no
	Sponsor    string // yes, no, unsure
	Salary     int64  // annual pounds, 0 = not provided
	Background string // none, uk_graduate, top_university
}

type MatchLevel string

const (
	MatchHigh   MatchLevel = "high"
	MatchMedium MatchLevel = "medium"
	MatchLow    MatchLevel = "low"
)

type Recommendation struct {
	Route   models.VisaRoute
	Match   MatchLevel
	Reasons []string
}

func Recommend(input Input, routes []models.VisaRoute) []Recommendation {
	bySlug := make(map[string]models.VisaRoute, len(routes))
	for _, r := range routes {
		bySlug[r.Slug] = r
	}

	var recs []Recommendation

	add := func(slug string, match MatchLevel, reasons ...string) {
		r, ok := bySlug[slug]
		if !ok {
			return
		}
		recs = append(recs, Recommendation{
			Route:   r,
			Match:   match,
			Reasons: reasons,
		})
	}

	hasOffer := input.JobOffer == "yes"
	hasSponsor := input.Sponsor == "yes"
	salary := input.Salary
	meetsSW := meetsRouteThreshold(bySlug["skilled-worker"], salary)
	meetsScaleUp := meetsRouteThreshold(bySlug["scale-up"], salary)
	meetsICT := salary > 0 && meetsRouteThreshold(bySlug["intra-company-transfer"], salary)

	switch {
	case !hasOffer && input.Background == "uk_graduate":
		add("graduate", MatchHigh,
			"You have a UK degree and no job offer yet. The Graduate route lets you work without a sponsor for 2 years.",
			"Often used as a bridge to Skilled Worker once you find a role.")
		add("skilled-worker", MatchLow,
			"You'll need a job offer and sponsor to switch before your Graduate visa expires.")

	case !hasOffer && input.Background == "top_university":
		add("high-potential-individual", MatchHigh,
			"Graduates of eligible top global universities can work in the UK for 2 years (3 years with a PhD) without a job offer or sponsor.",
			"Plan to switch to Skilled Worker or Global Talent before it expires. HPI cannot be extended.")
		add("global-talent", MatchMedium,
			"If you have exceptional talent in tech or STEM, endorsement may be possible without a job offer.")

	case !hasOffer:
		add("high-potential-individual", MatchMedium,
			"If you graduated from a Home Office listed top global university, you may qualify without a job offer.")
		add("global-talent", MatchMedium,
			"Leaders and emerging talent in tech can apply with endorsement. No job offer required.")
		add("graduate", MatchLow,
			"Only if you recently completed a degree at a UK university.")

	case hasOffer && hasSponsor:
		if salary == 0 || meetsSW {
			add("skilled-worker", MatchHigh,
				"You have a job offer and licensed sponsor. This is the main work visa route.",
				thresholdReason(bySlug["skilled-worker"], salary))
		} else {
			add("skilled-worker", MatchLow,
				"You have a sponsor, but your salary may be below the "+bySlug["skilled-worker"].ThresholdPounds()+" general threshold.",
				"Check the going rate for your SOC code. Some roles have different minimums.")
		}
		if salary == 0 || meetsScaleUp {
			add("scale-up", MatchMedium,
				"Qualifying high-growth companies can sponsor on a slightly lower threshold ("+bySlug["scale-up"].ThresholdPounds()+").",
				"Sponsor required for the first 6 months, then you can change employers freely.")
		}
		if meetsICT {
			add("intra-company-transfer", MatchMedium,
				"Suitable if your employer is transferring you from an overseas branch.",
				"Higher salary threshold ("+bySlug["intra-company-transfer"].ThresholdPounds()+"), typically for senior or specialist roles.")
		}

	case hasOffer && input.Sponsor == "unsure":
		add("skilled-worker", MatchMedium,
			"You have a job offer. Confirm your employer is on the Register of Licensed Sponsors.",
			"Use our sponsor search to check before applying.")
		if salary == 0 || meetsScaleUp {
			add("scale-up", MatchLow,
				"Worth exploring if your employer is a qualifying scale-up company.")
		}

	case hasOffer && !hasSponsor:
		add("global-talent", MatchHigh,
			"No sponsor needed. Requires endorsement from a designated body (e.g. for tech talent).",
			"Ideal if your employer won't sponsor but you have a strong track record.")
		add("skilled-worker", MatchLow,
			"You'll need to find a licensed sponsor willing to employ you on this route.")
	}

	// Deduplicate by slug, keeping highest match.
	recs = dedupeBest(recs)
	sortRecs(recs)
	return recs
}

func meetsRouteThreshold(r models.VisaRoute, salary int64) bool {
	if salary == 0 {
		return true
	}
	if r.SalaryThreshold <= 0 {
		return true
	}
	return salary >= r.SalaryThreshold/100
}

func thresholdReason(r models.VisaRoute, salary int64) string {
	if salary == 0 {
		return "Confirm your salary meets the " + r.ThresholdPounds() + "/yr threshold (or the going rate for your job code)."
	}
	if salary >= r.SalaryThreshold/100 {
		return "Your stated salary meets the general threshold."
	}
	return "Verify salary against the going rate for your specific SOC code."
}

func dedupeBest(recs []Recommendation) []Recommendation {
	best := make(map[string]Recommendation)
	order := []string{}
	for _, rec := range recs {
		existing, ok := best[rec.Route.Slug]
		if !ok || matchRank(rec.Match) > matchRank(existing.Match) {
			if !ok {
				order = append(order, rec.Route.Slug)
			}
			best[rec.Route.Slug] = rec
		}
	}
	out := make([]Recommendation, 0, len(best))
	for _, slug := range order {
		if rec, ok := best[slug]; ok {
			out = append(out, rec)
		}
	}
	return out
}

func matchRank(m MatchLevel) int {
	switch m {
	case MatchHigh:
		return 3
	case MatchMedium:
		return 2
	default:
		return 1
	}
}

func sortRecs(recs []Recommendation) {
	sort.Slice(recs, func(i, j int) bool {
		if matchRank(recs[i].Match) != matchRank(recs[j].Match) {
			return matchRank(recs[i].Match) > matchRank(recs[j].Match)
		}
		return recs[i].Route.Name < recs[j].Route.Name
	})
}
