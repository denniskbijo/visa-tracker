package models

import "time"

type Sponsor struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	City       string    `json:"city"`
	County     string    `json:"county"`
	Route      string    `json:"route"`
	Rating     string    `json:"rating"`
	SubRating  string    `json:"sub_rating"`
	IngestedAt time.Time `json:"ingested_at"`
}

type SponsorSearchResult struct {
	Sponsors   []Sponsor `json:"sponsors"`
	Total      int       `json:"total"`
	Query      string    `json:"query"`
	RouteFilter string   `json:"route_filter,omitempty"`
	CityFilter  string   `json:"city_filter,omitempty"`
}
