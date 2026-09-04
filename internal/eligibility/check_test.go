package eligibility

import (
	"testing"

	"github.com/denniskbijo/visa-tracker/internal/models"
)

func swRoute() models.VisaRoute {
	return models.VisaRoute{
		Slug:            "skilled-worker",
		Name:            "Skilled Worker",
		SalaryThreshold: 4170000,
	}
}

func TestCheck_noSalaryThreshold(t *testing.T) {
	res := Check(Input{
		Route:        models.VisaRoute{Slug: "graduate", SalaryThreshold: 0},
		SalaryPounds: 40000,
	})
	if res.Status != StatusNA {
		t.Fatalf("expected na, got %s", res.Status)
	}
}

func TestCheck_missingSalary(t *testing.T) {
	res := Check(Input{Route: swRoute()})
	if res.Status != StatusNA || res.Title != "Enter your salary" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestCheck_meetsGeneralThreshold(t *testing.T) {
	res := Check(Input{Route: swRoute(), SalaryPounds: 42000})
	if res.Status != StatusGreen {
		t.Fatalf("expected green, got %s", res.Status)
	}
}

func TestCheck_belowGoingRate(t *testing.T) {
	soc := &models.SOCCode{Code: "2134", GoingRatePence: 5470000}
	res := Check(Input{Route: swRoute(), SalaryPounds: 45000, SOC: soc})
	if res.Status != StatusRed {
		t.Fatalf("expected red, got %s (required %d)", res.Status, res.RequiredPence)
	}
	if res.RequiredPence != 5470000 {
		t.Fatalf("expected going rate as required, got %d", res.RequiredPence)
	}
}

func TestCheck_meetsGoingRate(t *testing.T) {
	soc := &models.SOCCode{Code: "2134", GoingRatePence: 5470000}
	res := Check(Input{Route: swRoute(), SalaryPounds: 54700, SOC: soc})
	if res.Status != StatusGreen {
		t.Fatalf("expected green, got %s", res.Status)
	}
}

func TestCheck_closeToThreshold(t *testing.T) {
	res := Check(Input{Route: swRoute(), SalaryPounds: 41000})
	if res.Status != StatusAmber {
		t.Fatalf("expected amber, got %s", res.Status)
	}
}

func TestCheck_islReducedThreshold(t *testing.T) {
	soc := &models.SOCCode{
		Code:                    "2112",
		GoingRatePence:          4030000,
		OnImmigrationSalaryList: true,
	}
	res := Check(Input{Route: swRoute(), SalaryPounds: 40300, SOC: soc})
	if res.Status != StatusAmber || res.Title != "May qualify via ISL" {
		t.Fatalf("expected ISL amber, got %+v", res)
	}
}

func TestCheck_islStillNeedsGoingRate(t *testing.T) {
	soc := &models.SOCCode{
		Code:                    "2112",
		GoingRatePence:          4030000,
		OnImmigrationSalaryList: true,
	}
	res := Check(Input{Route: swRoute(), SalaryPounds: 35000, SOC: soc})
	if res.Status != StatusRed {
		t.Fatalf("expected red below ISL going rate, got %+v", res)
	}
}

func TestCheck_unknownSOCUsesGeneral(t *testing.T) {
	res := Check(Input{Route: swRoute(), SalaryPounds: 42000, SOCCodeProvided: "9999"})
	if res.Status != StatusGreen {
		t.Fatalf("expected green with general threshold, got %s", res.Status)
	}
	if len(res.Notes) == 0 {
		t.Fatal("expected note about unknown SOC")
	}
}
