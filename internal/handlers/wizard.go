package handlers

import (
	"net/http"
	"strconv"

	"github.com/denniskbijo/visa-tracker/internal/wizard"
)

func (h *Handler) handleWizard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title     string
		ActiveNav string
	}{
		Title:     "Visa wizard",
		ActiveNav: "wizard",
	}
	h.render(w, "wizard.html", data)
}

func (h *Handler) handleWizardResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	salary, _ := strconv.ParseInt(r.FormValue("salary"), 10, 64)

	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		http.Error(w, "failed to load routes", http.StatusInternalServerError)
		return
	}

	input := wizard.Input{
		JobOffer:   r.FormValue("job_offer"),
		Sponsor:    r.FormValue("sponsor"),
		Salary:     salary,
		Background: r.FormValue("background"),
	}

	recs := wizard.Recommend(input, routes)

	data := struct {
		Recommendations []wizard.Recommendation
		Input           wizard.Input
	}{
		Recommendations: recs,
		Input:           input,
	}

	h.renderPartial(w, "wizard_results", data)
}
