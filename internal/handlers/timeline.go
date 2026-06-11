package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

type timelineRoute struct {
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	DurationYears float64 `json:"durationYears"`
	ILRYears      float64 `json:"ilrYears"`
	Extendable    bool    `json:"extendable"`
	Note          string  `json:"note"`
}

func timelineRoutesFromModels(routes []models.VisaRoute) []timelineRoute {
	out := make([]timelineRoute, 0, len(routes))
	for _, r := range routes {
		tr := timelineRoute{
			Slug:          r.Slug,
			Name:          r.Name,
			DurationYears: r.DurationYears,
		}
		switch r.Slug {
		case "skilled-worker", "scale-up", "intra-company-transfer":
			tr.ILRYears = 5
			tr.Extendable = true
		case "global-talent":
			tr.ILRYears = 3
			tr.Extendable = true
		case "graduate", "high-potential-individual":
			tr.ILRYears = 0
			tr.Extendable = false
			tr.Note = "Cannot be extended. Switch to Skilled Worker or Global Talent before expiry. ILR clock typically starts on a qualifying route."
		}
		out = append(out, tr)
	}
	return out
}

func (h *Handler) handleTimeline(w http.ResponseWriter, r *http.Request) {
	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		http.Error(w, "failed to load routes", http.StatusInternalServerError)
		return
	}

	tr := timelineRoutesFromModels(routes)
	jsonData, err := json.Marshal(tr)
	if err != nil {
		http.Error(w, "failed to encode routes", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		ActiveNav  string
		Routes     []models.VisaRoute
		RoutesJSON template.JS
	}{
		Title:      "My timeline",
		ActiveNav:  "timeline",
		Routes:     routes,
		RoutesJSON: template.JS(jsonData),
	}

	h.render(w, "timeline.html", data)
}
