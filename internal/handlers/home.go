package handlers

import "net/http"

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		http.Error(w, "failed to load visa routes", http.StatusInternalServerError)
		return
	}

	sponsorCount, _ := h.db.SponsorCount()
	socCodes, _ := h.db.AllSOCCodes()

	data := struct {
		Title        string
		ActiveNav    string
		Routes       interface{}
		RouteCount   int
		SponsorCount int
		SOCCount     int
	}{
		ActiveNav:    "home",
		Routes:       routes,
		RouteCount:   len(routes),
		SponsorCount: sponsorCount,
		SOCCount:     len(socCodes),
	}

	h.render(w, "home.html", data)
}
