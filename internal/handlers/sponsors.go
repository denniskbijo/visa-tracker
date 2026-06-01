package handlers

import (
	"net/http"
	"strconv"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (h *Handler) handleSponsors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	routeFilter := r.URL.Query().Get("route")
	cityFilter := r.URL.Query().Get("city")

	total, _ := h.db.SponsorCount()
	cities, _ := h.db.DistinctSponsorCities(30)

	var results *models.SponsorSearchResult
	var err error
	if q != "" || routeFilter != "" || cityFilter != "" {
		results, err = h.db.SearchSponsors(q, routeFilter, cityFilter, 50, 0)
		if err != nil {
			http.Error(w, "search failed", http.StatusInternalServerError)
			return
		}
	} else {
		results = &models.SponsorSearchResult{
			Query:       q,
			RouteFilter: routeFilter,
			CityFilter:  cityFilter,
		}
	}

	data := struct {
		Title         string
		ActiveNav     string
		TotalSponsors int
		Query         string
		RouteFilter   string
		CityFilter    string
		Cities        []string
		Results       interface{}
	}{
		Title:         "Sponsor Search",
		ActiveNav:     "sponsors",
		TotalSponsors: total,
		Query:         q,
		RouteFilter:   routeFilter,
		CityFilter:    cityFilter,
		Cities:        cities,
		Results:       results,
	}

	h.render(w, "sponsor_search.html", data)
}

func (h *Handler) handleSponsorResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	routeFilter := r.URL.Query().Get("route")
	cityFilter := r.URL.Query().Get("city")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if q == "" && routeFilter == "" && cityFilter == "" {
		h.renderPartial(w, "sponsor_results", &models.SponsorSearchResult{
			Query:       q,
			RouteFilter: routeFilter,
			CityFilter:  cityFilter,
		})
		return
	}

	results, err := h.db.SearchSponsors(q, routeFilter, cityFilter, 50, offset)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	h.renderPartial(w, "sponsor_results", results)
}
