package ingest

import (
	"fmt"
	"log"

	"github.com/denniskbijo/visa-tracker/internal/config"
	"github.com/denniskbijo/visa-tracker/internal/fsutil"
	"github.com/denniskbijo/visa-tracker/internal/models"
	"github.com/denniskbijo/visa-tracker/internal/store"
	"gopkg.in/yaml.v3"
)

type Ingester struct {
	db  *store.DB
	cfg *config.Config
}

func New(db *store.DB, cfg *config.Config) *Ingester {
	return &Ingester{db: db, cfg: cfg}
}

type thresholdsFile struct {
	VisaRoutes []struct {
		Slug                 string  `yaml:"slug"`
		Name                 string  `yaml:"name"`
		Description          string  `yaml:"description"`
		RequiresSponsor      bool    `yaml:"requires_sponsor"`
		RequiresEndorsement  bool    `yaml:"requires_endorsement"`
		SalaryThresholdPence int64   `yaml:"salary_threshold_pence"`
		DurationYears        float64 `yaml:"duration_years"`
	} `yaml:"visa_routes"`

	Thresholds []struct {
		VisaRouteSlug string `yaml:"visa_route_slug"`
		SOCCode       string `yaml:"soc_code"`
		AmountPence   int64  `yaml:"amount_pence"`
		EffectiveDate string `yaml:"effective_date"`
		Notes         string `yaml:"notes"`
	} `yaml:"thresholds"`

	ProcessingTimes []struct {
		VisaRouteSlug string `yaml:"visa_route_slug"`
		MedianDays    int    `yaml:"median_days"`
		P90Days       int    `yaml:"p90_days"`
		SnapshotDate  string `yaml:"snapshot_date"`
	} `yaml:"processing_times"`
}

type socFile struct {
	SOCCodes []struct {
		Code                    string `yaml:"code"`
		Title                   string `yaml:"title"`
		Description             string `yaml:"description"`
		GoingRatePence          int64  `yaml:"going_rate_pence"`
		OnImmigrationSalaryList bool   `yaml:"on_immigration_salary_list"`
	} `yaml:"soc_codes"`
}

func (ing *Ingester) LoadSeedData() error {
	if err := ing.loadThresholds(); err != nil {
		return fmt.Errorf("load thresholds: %w", err)
	}
	if err := ing.loadSOCCodes(); err != nil {
		return fmt.Errorf("load soc codes: %w", err)
	}
	return nil
}

func (ing *Ingester) loadThresholds() error {
	data, err := fsutil.ReadFileUnderRoot(ing.cfg.DataDir, "thresholds.yaml")
	if err != nil {
		return fmt.Errorf("read thresholds.yaml: %w", err)
	}

	var f thresholdsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("parse thresholds.yaml: %w", err)
	}

	for _, vr := range f.VisaRoutes {
		route := &models.VisaRoute{
			Slug:                vr.Slug,
			Name:                vr.Name,
			Description:         vr.Description,
			RequiresSponsor:     vr.RequiresSponsor,
			RequiresEndorsement: vr.RequiresEndorsement,
			SalaryThreshold:     vr.SalaryThresholdPence,
			DurationYears:       vr.DurationYears,
		}
		if err := ing.db.UpsertVisaRoute(route); err != nil {
			return fmt.Errorf("upsert visa route %s: %w", vr.Slug, err)
		}
	}
	log.Printf("loaded %d visa routes", len(f.VisaRoutes))

	for _, t := range f.Thresholds {
		route, err := ing.db.GetVisaRouteBySlug(t.VisaRouteSlug)
		if err != nil {
			return fmt.Errorf("lookup route %s: %w", t.VisaRouteSlug, err)
		}
		if route == nil {
			log.Printf("warning: unknown visa route slug %q in thresholds", t.VisaRouteSlug)
			continue
		}
		threshold := &models.SalaryThreshold{
			VisaRouteID:   route.ID,
			SOCCode:       t.SOCCode,
			AmountPence:   t.AmountPence,
			EffectiveDate: t.EffectiveDate,
			Notes:         t.Notes,
		}
		if err := ing.db.InsertThreshold(threshold); err != nil {
			return fmt.Errorf("insert threshold: %w", err)
		}
	}
	log.Printf("loaded %d salary thresholds", len(f.Thresholds))

	for _, pt := range f.ProcessingTimes {
		route, err := ing.db.GetVisaRouteBySlug(pt.VisaRouteSlug)
		if err != nil {
			return fmt.Errorf("lookup route %s: %w", pt.VisaRouteSlug, err)
		}
		if route == nil {
			continue
		}
		p := &models.ProcessingTime{
			VisaRouteID:  route.ID,
			MedianDays:   pt.MedianDays,
			P90Days:      pt.P90Days,
			SnapshotDate: pt.SnapshotDate,
		}
		if err := ing.db.InsertProcessingTime(p); err != nil {
			return fmt.Errorf("insert processing time: %w", err)
		}
	}
	log.Printf("loaded %d processing times", len(f.ProcessingTimes))

	return nil
}

func (ing *Ingester) loadSOCCodes() error {
	data, err := fsutil.ReadFileUnderRoot(ing.cfg.DataDir, "soc_codes.yaml")
	if err != nil {
		return fmt.Errorf("read soc_codes.yaml: %w", err)
	}

	var f socFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("parse soc_codes.yaml: %w", err)
	}

	for _, s := range f.SOCCodes {
		code := &models.SOCCode{
			Code:                    s.Code,
			Title:                   s.Title,
			Description:             s.Description,
			GoingRatePence:          s.GoingRatePence,
			OnImmigrationSalaryList: s.OnImmigrationSalaryList,
		}
		if err := ing.db.UpsertSOCCode(code); err != nil {
			return fmt.Errorf("upsert soc code %s: %w", s.Code, err)
		}
	}
	log.Printf("loaded %d SOC codes", len(f.SOCCodes))

	return nil
}
