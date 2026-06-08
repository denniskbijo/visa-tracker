package eligibility

import (
	"github.com/denniskbijo/visa-tracker/internal/models"
)

type Status string

const (
	StatusGreen Status = "green"
	StatusAmber Status = "amber"
	StatusRed   Status = "red"
	StatusNA    Status = "na"
)

const (
	islFloorPence    int64 = 3096000 // £30,960 ISL minimum
	islRatePercent   int64 = 80
	amberMarginPence int64 = 300000  // £3,000
	amberMarginPct   int64 = 10
)

type Input struct {
	Route             models.VisaRoute
	SalaryPounds      int64
	SOC               *models.SOCCode
	SOCCodeProvided   string
	RouteSOCThreshold *models.SalaryThreshold
}

type Result struct {
	Status         Status
	RequiredPence  int64
	ShortfallPence int64
	SalaryPounds   int64
	Title          string
	Message        string
	Notes          []string
}

func (r Result) RequiredPounds() string {
	return formatPounds(r.RequiredPence / 100)
}

func (r Result) ShortfallPounds() string {
	return formatPounds(r.ShortfallPence / 100)
}

func (r Result) SalaryDisplay() string {
	return formatPoundsPlain(r.SalaryPounds)
}

func Check(in Input) Result {
	if in.Route.SalaryThreshold == 0 {
		return Result{
			Status:  StatusNA,
			Title:   "No salary threshold",
			Message: "This route has no minimum salary requirement.",
		}
	}

	if in.SalaryPounds <= 0 {
		return Result{
			Status:  StatusNA,
			Title:   "Enter your salary",
			Message: "Add your annual salary to see how it compares to the threshold.",
		}
	}

	required := in.Route.SalaryThreshold
	var notes []string

	if in.SOC != nil {
		if in.SOC.GoingRatePence > required {
			required = in.SOC.GoingRatePence
		}
	} else if in.SOCCodeProvided != "" {
		notes = append(notes, "SOC code not found in our database. Using the general threshold only.")
	} else {
		notes = append(notes, "Add your SOC code to check against the going rate for your job.")
	}

	if in.RouteSOCThreshold != nil && in.RouteSOCThreshold.AmountPence > required {
		required = in.RouteSOCThreshold.AmountPence
	}

	salaryPence := in.SalaryPounds * 100
	shortfall := required - salaryPence
	if shortfall < 0 {
		shortfall = 0
	}

	res := Result{
		RequiredPence:  required,
		ShortfallPence: shortfall,
		SalaryPounds:   in.SalaryPounds,
		Notes:          notes,
	}

	if salaryPence >= required {
		res.Status = StatusGreen
		res.Title = "Meets threshold"
		res.Message = "Your salary meets the required minimum for this route."
		return res
	}

	if in.Route.Slug == "skilled-worker" && in.SOC != nil && in.SOC.OnImmigrationSalaryList {
		islRequired := islFloorPence
		if reduced := in.SOC.GoingRatePence * islRatePercent / 100; reduced > islRequired {
			islRequired = reduced
		}
		if salaryPence >= islRequired {
			res.Status = StatusAmber
			res.Title = "May qualify via ISL"
			res.Message = "Your salary is below the standard going rate but may qualify under the Immigration Salary List at a reduced threshold."
			res.Notes = append(res.Notes, "Confirm ISL eligibility and exact rules on gov.uk before applying.")
			return res
		}
	}

	if shortfall <= amberMarginPence || (required > 0 && shortfall*100/required <= amberMarginPct) {
		res.Status = StatusAmber
		res.Title = "Close to threshold"
		res.Message = "Your salary is slightly below the required minimum. A small increase may bring you within range."
		return res
	}

	res.Status = StatusRed
	res.Title = "Below threshold"
	res.Message = "Your salary is below the required minimum for this route."
	return res
}
