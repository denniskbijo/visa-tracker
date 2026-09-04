package handlers

import (
	"net/http"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

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

	var skilledWorker *models.VisaRoute
	var skilledWorkerNote string
	if sw, _ := h.db.GetVisaRouteBySlug("skilled-worker"); sw != nil {
		skilledWorker = sw
		if t, _ := h.db.LatestGeneralThreshold(sw.ID); t != nil {
			skilledWorkerNote = t.Notes
			if t.EffectiveDate != "" {
				skilledWorkerNote = t.Notes + " · effective " + t.EffectiveDate
			}
		}
	}

	data := struct {
		Title             string
		ActiveNav         string
		Routes            interface{}
		RouteCount        int
		SponsorCount      int
		SOCCount          int
		SkilledWorker     *models.VisaRoute
		SkilledWorkerNote string
	}{
		ActiveNav:         "home",
		Routes:            routes,
		RouteCount:        len(routes),
		SponsorCount:      sponsorCount,
		SOCCount:          len(socCodes),
		SkilledWorker:     skilledWorker,
		SkilledWorkerNote: skilledWorkerNote,
	}

	h.render(w, "home.html", data)
}
