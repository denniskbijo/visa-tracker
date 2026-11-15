package wizard

import (
	"sort"
	"strconv"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

// SkilledWorkerILRYears is the settlement count the timeline already uses for
// Skilled Worker. The quiz points at that timeline. It does not add a new rule.
const SkilledWorkerILRYears = 5

type Input struct {
	OnSkilledWorker string // yes, not_yet
	YearsOnRoute    string // under_3, 3_to_5, 5_plus
	NextStep        string // stay, global_talent, scale_up
	OfferSponsor    string // sponsor_yes, sponsor_unsure, sponsor_no, no_offer
	Salary          int64  // annual pounds, 0 = not provided
	Background      string // uk_graduate, top_university, tech_talent, none
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
	b := &builder{bySlug: bySlug}
	if input.OnSkilledWorker == "yes" {
		b.currentSkilledWorker(input)
	} else {
		b.notYet(input)
	}
	recs := dedupeBest(b.recs)
	sortRecs(recs)
	return recs
}

type builder struct {
	bySlug map[string]models.VisaRoute
	recs   []Recommendation
}

func (b *builder) add(slug string, match MatchLevel, situational ...string) {
	r, ok := b.bySlug[slug]
	if !ok {
		return
	}
	reasons := make([]string, 0, len(situational)+1)
	for _, s := range situational {
		if s != "" {
			reasons = append(reasons, s)
		}
	}
	if desc := cleanDesc(r.Description); desc != "" {
		reasons = append(reasons, desc)
	}
	b.recs = append(b.recs, Recommendation{
		Route:   r,
		Match:   match,
		Reasons: reasons,
	})
}

func (b *builder) currentSkilledWorker(input Input) {
	sw := b.bySlug["skilled-worker"]
	b.add("skilled-worker", MatchHigh,
		ilrReadiness(input.YearsOnRoute),
		salaryCompare(sw, input.Salary),
		"You already have a sponsored UK job on this route. The Skilled Worker page checks your salary against the general threshold and your job code.",
	)

	switch input.NextStep {
	case "global_talent":
		gt := b.bySlug["global-talent"]
		b.add("global-talent", MatchMedium,
			"A switch to look at if you have a strong tech or STEM record. No sponsor is required.",
			salaryCompare(gt, input.Salary),
		)
	case "scale_up":
		su := b.bySlug["scale-up"]
		b.add("scale-up", salaryMatch(MatchMedium, su, input.Salary),
			"A switch only if your employer is a qualifying high-growth sponsor.",
			salaryCompare(su, input.Salary),
		)
	}
}

func (b *builder) notYet(input Input) {
	switch input.OfferSponsor {
	case "sponsor_yes":
		b.offerWithSponsor(input)
	case "sponsor_unsure":
		b.offerUnsure(input)
	case "sponsor_no":
		b.offerWithoutSponsor(input)
	default:
		b.noOffer(input)
	}
}

func (b *builder) offerWithSponsor(input Input) {
	sw := b.bySlug["skilled-worker"]
	b.add("skilled-worker", salaryMatch(MatchHigh, sw, input.Salary),
		"You have a UK tech or STEM offer from a licensed sponsor. Skilled Worker is the usual sponsored route.",
		salaryCompare(sw, input.Salary),
	)

	su := b.bySlug["scale-up"]
	b.add("scale-up", salaryMatch(MatchMedium, su, input.Salary),
		"Also possible if the employer is a qualifying high-growth sponsor.",
		salaryCompare(su, input.Salary),
	)

	// ICT is a transfer route, not a settlement route. Show it only when the
	// stated salary meets the stored threshold, and say it does not lead to ILR.
	ict := b.bySlug["intra-company-transfer"]
	if input.Salary > 0 && meetsRouteThreshold(ict, input.Salary) {
		b.add("intra-company-transfer", MatchLow,
			"Only if a group company is transferring you from an overseas office to a UK branch.",
			"Senior or Specialist Worker (ICT) does not lead to indefinite leave to remain (ILR). Time on this route does not count toward settlement.",
			salaryCompare(ict, input.Salary),
		)
	}
}

func (b *builder) offerUnsure(input Input) {
	sw := b.bySlug["skilled-worker"]
	b.add("skilled-worker", salaryMatch(MatchMedium, sw, input.Salary),
		"You have a UK job offer. Confirm the employer is on the Register of Licensed Sponsors before you rely on Skilled Worker.",
		salaryCompare(sw, input.Salary),
	)
	su := b.bySlug["scale-up"]
	b.add("scale-up", salaryMatch(MatchLow, su, input.Salary),
		"Worth a look if the employer is a qualifying high-growth sponsor. Check the sponsor register first.",
		salaryCompare(su, input.Salary),
	)
}

func (b *builder) offerWithoutSponsor(input Input) {
	b.unsponsored(input, true)
	sw := b.bySlug["skilled-worker"]
	b.add("skilled-worker", MatchLow,
		"Skilled Worker needs a licensed sponsor. This offer does not give you that route unless the employer gets a licence, or you find one who has.",
		salaryCompare(sw, input.Salary),
	)
}

func (b *builder) noOffer(input Input) {
	b.unsponsored(input, false)
}

// unsponsored recommends routes that do not need a sponsor.
// withOffer is true when they have a job the employer will not sponsor.
func (b *builder) unsponsored(input Input, withOffer bool) {
	switch input.Background {
	case "uk_graduate":
		b.add("graduate", MatchHigh,
			graduateFit(withOffer),
		)
		if !withOffer {
			sw := b.bySlug["skilled-worker"]
			b.add("skilled-worker", MatchLow,
				"To move on from Graduate you need a job offer from a licensed sponsor.",
				salaryCompare(sw, input.Salary),
			)
		}
	case "top_university":
		b.add("high-potential-individual", MatchHigh,
			"Your degree may qualify without a job offer or a sponsor, if the university is on the Home Office list.",
		)
		b.add("global-talent", MatchMedium,
			"If you also have a strong tech or STEM record, Global Talent is the endorsement route and does not need a sponsor.",
		)
	case "tech_talent":
		b.add("global-talent", MatchHigh,
			"A strong tech or STEM record can be endorsed without a job offer or a sponsor.",
		)
		b.add("high-potential-individual", MatchMedium,
			"Also check the Home Office list if your degree is from a listed university. That route does not need a sponsor.",
		)
	default:
		b.add("global-talent", MatchMedium,
			"Leaders and emerging talent in tech can apply with an endorsement. No job offer or sponsor is required.",
		)
		b.add("high-potential-individual", MatchMedium,
			"Possible if you graduated from a university on the Home Office list in the last 5 years. No sponsor is required.",
		)
		b.add("graduate", MatchLow,
			"Only if you recently finished a degree at a UK university.",
		)
	}
}

func graduateFit(withOffer bool) string {
	if withOffer {
		return "The employer will not sponsor, but a recent UK degree can still let you take the job on the Graduate route. No sponsor or salary floor in our data."
	}
	return "You have a UK degree and no sponsored job yet. Graduate lets you work in tech or STEM without a sponsor or a salary floor."
}

func ilrReadiness(years string) string {
	span := strconv.Itoa(SkilledWorkerILRYears)
	switch years {
	case "under_3":
		return "You are early on Skilled Worker. This site's timeline counts " + span + " years on this route toward indefinite leave to remain (ILR). Save your visa start date there."
	case "3_to_5":
		return "You are in the later part of the " + span + "-year ILR count this site uses for Skilled Worker. Check the date on My timeline."
	case "5_plus":
		return "You may already be at the " + span + "-year mark this site uses for Skilled Worker ILR. Check your start date on My timeline, then confirm the settlement rules on gov.uk before you apply."
	default:
		return "This site's timeline counts " + span + " years on Skilled Worker toward indefinite leave to remain (ILR). Add your start date there."
	}
}

func salaryCompare(r models.VisaRoute, salary int64) string {
	if r.SalaryThreshold <= 0 {
		return "Our data has no salary floor for this route."
	}
	threshold := r.ThresholdPounds()
	if salary <= 0 {
		return "Confirm your salary meets the " + threshold + " general threshold, or the going rate for your job code if that is higher."
	}
	if meetsRouteThreshold(r, salary) {
		return "Your stated salary meets the " + threshold + " general threshold. The going rate for your job code can be higher, so check it on the route page."
	}
	return "Your stated salary is below the " + threshold + " general threshold. The going rate for your job code can differ, so check both on the route page."
}

func salaryMatch(base MatchLevel, r models.VisaRoute, salary int64) MatchLevel {
	if salary > 0 && r.SalaryThreshold > 0 && !meetsRouteThreshold(r, salary) {
		return MatchLow
	}
	return base
}

func meetsRouteThreshold(r models.VisaRoute, salary int64) bool {
	if salary <= 0 || r.SalaryThreshold <= 0 {
		return true
	}
	return salary >= r.SalaryThreshold/100
}

func cleanDesc(s string) string {
	return strings.Join(strings.Fields(s), " ")
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
