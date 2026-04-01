package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) handleAPIVisas(w http.ResponseWriter, r *http.Request) {
	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		jsonError(w, "failed to load visa routes", http.StatusInternalServerError)
		return
	}
	jsonOK(w, routes)
}

func (h *Handler) handleAPISponsors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	routeFilter := r.URL.Query().Get("route")
	cityFilter := r.URL.Query().Get("city")

	results, err := h.db.SearchSponsors(q, routeFilter, cityFilter, 50, 0)
	if err != nil {
		jsonError(w, "search failed", http.StatusInternalServerError)
		return
	}
	jsonOK(w, results)
}

func (h *Handler) handleAPISOC(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	if q != "" {
		codes, err := h.db.SearchSOCCodes(q)
		if err != nil {
			jsonError(w, "search failed", http.StatusInternalServerError)
			return
		}
		jsonOK(w, codes)
		return
	}

	codes, err := h.db.AllSOCCodes()
	if err != nil {
		jsonError(w, "failed to load SOC codes", http.StatusInternalServerError)
		return
	}
	jsonOK(w, codes)
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode (ok): %v", err)
	}
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		log.Printf("json encode (error): %v", err)
	}
}
