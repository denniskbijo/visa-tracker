package models

import "time"

type VisaRoute struct {
	ID                  int64     `json:"id"`
	Slug                string    `json:"slug"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	RequiresSponsor     bool      `json:"requires_sponsor"`
	RequiresEndorsement bool      `json:"requires_endorsement"`
	SalaryThreshold     int64     `json:"salary_threshold"` // pence
	DurationYears       float64   `json:"duration_years"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (v VisaRoute) ThresholdPounds() string {
	if v.SalaryThreshold == 0 {
		return "None"
	}
	pounds := v.SalaryThreshold / 100
	return formatPounds(pounds)
}

func formatPounds(p int64) string {
	s := ""
	n := p
	for n > 0 {
		if s != "" {
			s = "," + s
		}
		chunk := n % 1000
		n /= 1000
		if n > 0 {
			// zero-pad
			if chunk < 10 {
				s = "00" + itoa(chunk) + s
			} else if chunk < 100 {
				s = "0" + itoa(chunk) + s
			} else {
				s = itoa(chunk) + s
			}
		} else {
			s = itoa(chunk) + s
		}
	}
	if s == "" {
		s = "0"
	}
	return "£" + s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

type SalaryThreshold struct {
	ID            int64     `json:"id"`
	VisaRouteID   int64     `json:"visa_route_id"`
	SOCCode       string    `json:"soc_code"`
	AmountPence   int64     `json:"amount_pence"`
	EffectiveDate string    `json:"effective_date"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t SalaryThreshold) AmountPounds() string {
	return formatPounds(t.AmountPence / 100)
}
