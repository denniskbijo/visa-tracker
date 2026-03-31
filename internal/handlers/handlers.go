package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/config"
	"github.com/denniskbijo/visa-tracker/internal/models"
	"github.com/denniskbijo/visa-tracker/internal/store"
)

type Handler struct {
	db   *store.DB
	cfg  *config.Config
	tmpl map[string]*template.Template
}

var funcMap = template.FuncMap{
	"upper":         strings.ToUpper,
	"threshold_bar": thresholdBar,
	"route_color":   routeColor,
}

func New(db *store.DB, cfg *config.Config) *Handler {
	h := &Handler{db: db, cfg: cfg}
	h.loadTemplates()
	return h
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(h.cfg.StaticDir))))

	mux.HandleFunc("/", h.handleHome)
	mux.HandleFunc("/visas", h.handleVisas)
	mux.HandleFunc("/visas/", h.handleVisaDetail)
	mux.HandleFunc("/sponsors", h.handleSponsors)
	mux.HandleFunc("/sponsors/results", h.handleSponsorResults)
	mux.HandleFunc("/soc", h.handleSOC)
	mux.HandleFunc("/soc/results", h.handleSOCResults)

	mux.HandleFunc("/api/v1/visas", h.handleAPIVisas)
	mux.HandleFunc("/api/v1/sponsors", h.handleAPISponsors)
	mux.HandleFunc("/api/v1/soc", h.handleAPISOC)
}

func (h *Handler) loadTemplates() {
	h.tmpl = make(map[string]*template.Template)
	layoutPath := filepath.Join(h.cfg.TemplatesDir, "layout.html")
	partialsGlob := filepath.Join(h.cfg.TemplatesDir, "partials", "*.html")

	pages := []string{
		"home.html",
		"visa_tracker.html",
		"visa_detail.html",
		"sponsor_search.html",
		"soc_lookup.html",
	}

	for _, page := range pages {
		pagePath := filepath.Join(h.cfg.TemplatesDir, page)
		t, err := template.New("").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
		if err != nil {
			log.Fatalf("parse template %s: %v", page, err)
		}
		t, err = t.ParseGlob(partialsGlob)
		if err != nil {
			log.Fatalf("parse partials for %s: %v", page, err)
		}
		h.tmpl[page] = t
	}

	// standalone partials for HTMX responses
	for _, partial := range []string{"sponsor_results.html", "soc_results.html"} {
		partialPath := filepath.Join(h.cfg.TemplatesDir, "partials", partial)
		t, err := template.New("").Funcs(funcMap).ParseFiles(partialPath)
		if err != nil {
			log.Fatalf("parse partial %s: %v", partial, err)
		}
		h.tmpl["partial:"+partial] = t
	}
}

func (h *Handler) render(w http.ResponseWriter, page string, data interface{}) {
	t, ok := h.tmpl[page]
	if !ok {
		log.Printf("template not found: %s", page)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("template error (%s): %v", page, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data interface{}) {
	t, ok := h.tmpl["partial:"+name+".html"]
	if !ok {
		log.Printf("partial template not found: %s", name)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("partial template error (%s): %v", name, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func thresholdBar(thresholdPence int64) int {
	if thresholdPence <= 0 {
		return 5
	}
	pct := int(thresholdPence * 100 / 5000000)
	if pct > 100 {
		pct = 100
	}
	if pct < 5 {
		pct = 5
	}
	return pct
}

func routeColor(slug string) string {
	colors := map[string]string{
		"skilled-worker":           "#1D9E75",
		"global-talent":            "#378ADD",
		"high-potential-individual": "#EF9F27",
		"scale-up":                 "#5DCAA5",
		"intra-company-transfer":   "#D4537E",
		"graduate":                 "#85B7EB",
	}
	if c, ok := colors[slug]; ok {
		return c
	}
	return "#8b949e"
}

type routeWithDetails struct {
	Route          models.VisaRoute
	ProcessingTime *models.ProcessingTime
	Thresholds     []models.SalaryThreshold
}
