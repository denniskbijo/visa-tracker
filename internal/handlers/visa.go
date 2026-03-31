package handlers

import (
	"net/http"
	"strings"
)

func (h *Handler) handleVisas(w http.ResponseWriter, r *http.Request) {
	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		http.Error(w, "failed to load visa routes", http.StatusInternalServerError)
		return
	}

	var details []routeWithDetails
	for _, route := range routes {
		pt, _ := h.db.LatestProcessingTime(route.ID)
		thresholds, _ := h.db.LatestThresholds(route.ID)
		details = append(details, routeWithDetails{
			Route:          route,
			ProcessingTime: pt,
			Thresholds:     thresholds,
		})
	}

	data := struct {
		Title             string
		ActiveNav         string
		Routes            interface{}
		RoutesWithDetails []routeWithDetails
	}{
		Title:             "Visa Dashboard",
		ActiveNav:         "visas",
		Routes:            routes,
		RoutesWithDetails: details,
	}

	h.render(w, "visa_tracker.html", data)
}

func (h *Handler) handleVisaDetail(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/visas/")
	if slug == "" {
		h.handleVisas(w, r)
		return
	}

	route, err := h.db.GetVisaRouteBySlug(slug)
	if err != nil {
		http.Error(w, "failed to load visa route", http.StatusInternalServerError)
		return
	}
	if route == nil {
		http.NotFound(w, r)
		return
	}

	pt, _ := h.db.LatestProcessingTime(route.ID)
	thresholds, _ := h.db.LatestThresholds(route.ID)

	data := struct {
		Title          string
		ActiveNav      string
		Route          interface{}
		ProcessingTime interface{}
		Thresholds     interface{}
	}{
		Title:          route.Name,
		ActiveNav:      "visas",
		Route:          route,
		ProcessingTime: pt,
		Thresholds:     thresholds,
	}

	h.render(w, "visa_detail.html", data)
}
