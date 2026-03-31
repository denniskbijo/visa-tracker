package handlers

import (
	"net/http"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func (h *Handler) handleSOC(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	allCodes, err := h.db.AllSOCCodes()
	if err != nil {
		http.Error(w, "failed to load SOC codes", http.StatusInternalServerError)
		return
	}

	codes := allCodes
	if q != "" {
		codes, err = h.db.SearchSOCCodes(q)
		if err != nil {
			http.Error(w, "search failed", http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		Title     string
		ActiveNav string
		AllCodes  []models.SOCCode
		Codes     []models.SOCCode
		Query     string
	}{
		Title:     "SOC Code Lookup",
		ActiveNav: "soc",
		AllCodes:  allCodes,
		Codes:     codes,
		Query:     q,
	}

	h.render(w, "soc_lookup.html", data)
}

func (h *Handler) handleSOCResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var codes []models.SOCCode
	var err error
	if q != "" {
		codes, err = h.db.SearchSOCCodes(q)
	} else {
		codes, err = h.db.AllSOCCodes()
	}
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	data := struct {
		Codes []models.SOCCode
		Query string
	}{
		Codes: codes,
		Query: q,
	}

	h.renderPartial(w, "soc_results", data)
}
