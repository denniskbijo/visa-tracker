package models

import "time"

type SOCCode struct {
	Code                    string    `json:"code"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	GoingRatePence          int64     `json:"going_rate_pence"`
	OnImmigrationSalaryList bool      `json:"on_immigration_salary_list"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (s SOCCode) GoingRatePounds() string {
	if s.GoingRatePence == 0 {
		return "N/A"
	}
	return formatPounds(s.GoingRatePence / 100)
}
