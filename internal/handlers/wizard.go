package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/wizard"
)

func (h *Handler) handleWizard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title           string
		ActiveNav       string
		MetaDescription string
		ILRYears        int
	}{
		Title:           "Take the UK visa quiz",
		ActiveNav:       "quiz",
		MetaDescription: "Four questions for UK tech and STEM professionals. See Skilled Worker settlement timing and salary, or which work route may fit if you are not sponsored yet.",
		ILRYears:        wizard.SkilledWorkerILRYears,
	}
	h.render(w, "wizard.html", data)
}

func (h *Handler) redirectQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	target := "/quiz"
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
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

	routes, err := h.db.AllVisaRoutes()
	if err != nil {
		http.Error(w, "failed to load routes", http.StatusInternalServerError)
		return
	}

	input := quizInput(r)
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

func quizInput(r *http.Request) wizard.Input {
	salary, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("salary")), 10, 64)
	if salary < 0 {
		salary = 0
	}
	return wizard.Input{
		OnSkilledWorker: r.FormValue("on_skilled_worker"),
		YearsOnRoute:    r.FormValue("years_on_route"),
		NextStep:        r.FormValue("next_step"),
		OfferSponsor:    r.FormValue("offer_sponsor"),
		Salary:          salary,
		Background:      r.FormValue("background"),
	}
}
