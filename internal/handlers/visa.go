package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/eligibility"
	"github.com/denniskbijo/visa-tracker/internal/models"
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
	rest := strings.TrimPrefix(r.URL.Path, "/visas/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		h.handleVisas(w, r)
		return
	}

	slug := parts[0]
	if len(parts) == 2 && parts[1] == "check" {
		h.handleEligibilityCheck(w, r, slug)
		return
	}
	if len(parts) > 1 {
		http.NotFound(w, r)
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
	socCodes, _ := h.db.AllSOCCodes()

	data := struct {
		Title          string
		ActiveNav      string
		Route          interface{}
		ProcessingTime interface{}
		Thresholds     interface{}
		SOCCodes       []models.SOCCode
	}{
		Title:          route.Name,
		ActiveNav:      "visas",
		Route:          route,
		ProcessingTime: pt,
		Thresholds:     thresholds,
		SOCCodes:       socCodes,
	}

	h.render(w, "visa_detail.html", data)
}

func (h *Handler) handleEligibilityCheck(w http.ResponseWriter, r *http.Request, slug string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

	salary, _ := strconv.ParseInt(r.URL.Query().Get("salary"), 10, 64)
	socCode := strings.TrimSpace(r.URL.Query().Get("soc"))

	var soc *models.SOCCode
	if socCode != "" {
		soc, err = h.db.GetSOCCode(socCode)
		if err != nil {
			http.Error(w, "failed to look up SOC code", http.StatusInternalServerError)
			return
		}
	}

	var routeSOCThreshold *models.SalaryThreshold
	if socCode != "" {
		routeSOCThreshold, err = h.db.LatestThresholdForSOC(route.ID, socCode)
		if err != nil {
			http.Error(w, "failed to load threshold", http.StatusInternalServerError)
			return
		}
	}

	result := eligibility.Check(eligibility.Input{
		Route:             *route,
		SalaryPounds:      salary,
		SOC:               soc,
		SOCCodeProvided:   socCode,
		RouteSOCThreshold: routeSOCThreshold,
	})

	data := struct {
		Result eligibility.Result
		Route  models.VisaRoute
		SOC    *models.SOCCode
	}{
		Result: result,
		Route:  *route,
		SOC:    soc,
	}

	h.renderPartial(w, "eligibility_result", data)
}
