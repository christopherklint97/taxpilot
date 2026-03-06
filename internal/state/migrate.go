package state

import (
	"fmt"
	"math"
)

// PriorYearContext holds extracted prior-year data relevant to the current year.
type PriorYearContext struct {
	PriorYear      int                `json:"prior_year"`
	FederalAGI     float64            `json:"federal_agi"`      // 1040 line 11
	CAAdjustedAGI  float64            `json:"ca_agi"`           // CA 540 line 17
	FilingStatus   string             `json:"filing_status"`
	FirstName      string             `json:"first_name"`
	LastName       string             `json:"last_name"`
	SSN            string             `json:"ssn"`
	TotalWages     float64            `json:"total_wages"`      // 1040 line 1a
	FedWithholding float64            `json:"fed_withholding"`  // 1040 line 25d
	CAWithholding  float64            `json:"ca_withholding"`   // CA 540 line 71
	FedTax         float64            `json:"fed_tax"`          // 1040 line 16
	CATax          float64            `json:"ca_tax"`           // CA 540 line 40
	FedRefund      float64            `json:"fed_refund"`       // 1040 line 34
	CARefund       float64            `json:"ca_refund"`        // CA 540 line 91
	// Prior-year CA AGI is critical for FTB e-file authentication
	PriorYearCAAGI float64            `json:"prior_year_ca_agi"`
	// All extracted values for reference
	AllValues    map[string]float64 `json:"all_values"`
	AllStrValues map[string]string  `json:"all_str_values"`
}

// CarryoverFields defines which fields carry over from prior year to current year.
// Key is the field key in prior year, value is description of why it carries over.
var CarryoverFields = map[string]string{
	"1040:filing_status": "Filing status usually stays the same",
	"1040:first_name":    "Personal info carries over",
	"1040:last_name":     "Personal info carries over",
	"1040:ssn":           "SSN never changes",
	"w2:1:employer_name": "Employer info likely same if still employed",
	"w2:1:employer_ein":  "Employer EIN",
}

// SignificantChangeThreshold defines what constitutes a "significant" change
// in a numeric field value (percentage).
const SignificantChangeThreshold = 0.20 // 20%

// knownNumericMappings maps well-known field keys to PriorYearContext fields.
var knownNumericMappings = map[string]func(*PriorYearContext) *float64{
	"1040:11":    func(c *PriorYearContext) *float64 { return &c.FederalAGI },
	"ca540:17":   func(c *PriorYearContext) *float64 { return &c.CAAdjustedAGI },
	"1040:1a":    func(c *PriorYearContext) *float64 { return &c.TotalWages },
	"1040:25d":   func(c *PriorYearContext) *float64 { return &c.FedWithholding },
	"ca540:71":   func(c *PriorYearContext) *float64 { return &c.CAWithholding },
	"1040:16":    func(c *PriorYearContext) *float64 { return &c.FedTax },
	"ca540:40":   func(c *PriorYearContext) *float64 { return &c.CATax },
	"1040:34":    func(c *PriorYearContext) *float64 { return &c.FedRefund },
	"ca540:91":   func(c *PriorYearContext) *float64 { return &c.CARefund },
}

// knownStringMappings maps well-known field keys to PriorYearContext string fields.
var knownStringMappings = map[string]func(*PriorYearContext) *string{
	"1040:filing_status": func(c *PriorYearContext) *string { return &c.FilingStatus },
	"1040:first_name":    func(c *PriorYearContext) *string { return &c.FirstName },
	"1040:last_name":     func(c *PriorYearContext) *string { return &c.LastName },
	"1040:ssn":           func(c *PriorYearContext) *string { return &c.SSN },
}

// ExtractPriorYearContext builds a PriorYearContext from a TaxReturn's
// computed values (either loaded from state or parsed from PDF).
func ExtractPriorYearContext(ret *TaxReturn) *PriorYearContext {
	if ret == nil {
		return &PriorYearContext{
			AllValues:    make(map[string]float64),
			AllStrValues: make(map[string]string),
		}
	}

	ctx := &PriorYearContext{
		PriorYear:    ret.TaxYear,
		AllValues:    make(map[string]float64),
		AllStrValues: make(map[string]string),
	}

	// Copy all numeric values (merge Inputs and Computed).
	for k, v := range ret.Inputs {
		ctx.AllValues[k] = v
	}
	for k, v := range ret.Computed {
		ctx.AllValues[k] = v
	}

	// Copy all string values.
	for k, v := range ret.StrInputs {
		ctx.AllStrValues[k] = v
	}

	// Extract well-known numeric fields.
	for key, setter := range knownNumericMappings {
		if v, ok := ctx.AllValues[key]; ok {
			*setter(ctx) = v
		}
	}

	// Extract well-known string fields.
	for key, setter := range knownStringMappings {
		if v, ok := ctx.AllStrValues[key]; ok {
			*setter(ctx) = v
		}
	}

	// Also check filing status in Inputs (some returns store it as string input).
	if ctx.FilingStatus == "" {
		if v, ok := ret.StrInputs["1040:filing_status"]; ok {
			ctx.FilingStatus = v
		}
	}

	// PriorYearCAAGI is the CA AGI from this return (will be "prior year" when used next year).
	ctx.PriorYearCAAGI = ctx.CAAdjustedAGI

	return ctx
}

// MigrateToCurrentYear takes a prior-year TaxReturn and creates a new
// TaxReturn for the current year with carried-over fields pre-filled.
func MigrateToCurrentYear(prior *TaxReturn, currentYear int) *TaxReturn {
	ret := NewTaxReturn(currentYear, prior.State)

	// Extract prior-year context and attach it.
	priorCtx := ExtractPriorYearContext(prior)
	ret.PriorYearCtx = priorCtx

	// Carry over string fields defined in CarryoverFields.
	for key := range CarryoverFields {
		if v, ok := prior.StrInputs[key]; ok {
			ret.StrInputs[key] = v
		}
	}

	// Carry over filing status to the top-level field as well.
	if fs, ok := prior.StrInputs["1040:filing_status"]; ok {
		ret.FilingStatus = fs
	} else if prior.FilingStatus != "" {
		ret.FilingStatus = prior.FilingStatus
	}

	// Copy numeric carryover fields (e.g., employer EIN as numeric).
	for key := range CarryoverFields {
		if v, ok := prior.Inputs[key]; ok {
			ret.Inputs[key] = v
		}
	}

	// Store prior-year values for reference in PriorYear map.
	for k, v := range priorCtx.AllValues {
		ret.PriorYear[k] = v
	}

	return ret
}

// ChangeFlag represents a significant difference between years.
type ChangeFlag struct {
	FieldKey      string  // e.g., "1040:1a"
	Label         string  // human-readable label
	PriorValue    float64
	CurrentValue  float64
	PercentChange float64
	Severity      string // "info", "warning", "attention"
	Message       string // explanation of why this might matter
}

// changeLabelMap provides human-readable labels for well-known field keys.
var changeLabelMap = map[string]string{
	"1040:11":  "Federal AGI",
	"1040:1a":  "Total Wages",
	"1040:25d": "Federal Withholding",
	"1040:16":  "Federal Tax",
	"1040:34":  "Federal Refund",
	"ca540:17": "CA Adjusted AGI",
	"ca540:71": "CA Withholding",
	"ca540:40": "CA Tax",
	"ca540:91": "CA Refund",
}

// CompareReturns compares a prior-year and current-year return, returning
// a list of significant changes that should be flagged to the user.
func CompareReturns(prior, current *PriorYearContext) []ChangeFlag {
	if prior == nil || current == nil {
		return nil
	}

	var flags []ChangeFlag

	// Compare all fields present in both contexts.
	for key, priorVal := range prior.AllValues {
		currentVal, ok := current.AllValues[key]
		if !ok {
			continue
		}

		pctChange := percentChange(priorVal, currentVal)
		if math.Abs(pctChange) < SignificantChangeThreshold {
			continue
		}

		label := changeLabelMap[key]
		if label == "" {
			label = key
		}

		severity := classifySeverity(pctChange)
		message := buildChangeMessage(label, priorVal, currentVal, pctChange)

		flags = append(flags, ChangeFlag{
			FieldKey:      key,
			Label:         label,
			PriorValue:    priorVal,
			CurrentValue:  currentVal,
			PercentChange: pctChange,
			Severity:      severity,
			Message:       message,
		})
	}

	return flags
}

// percentChange computes (current - prior) / |prior|.
// Returns 0 if prior is zero and current is also zero.
// Returns 1.0 (100%) if prior is zero but current is non-zero.
func percentChange(prior, current float64) float64 {
	if prior == 0 {
		if current == 0 {
			return 0
		}
		return 1.0 // 100% change from zero
	}
	return (current - prior) / math.Abs(prior)
}

// classifySeverity returns severity based on the magnitude of change.
func classifySeverity(pctChange float64) string {
	abs := math.Abs(pctChange)
	switch {
	case abs >= 0.50:
		return "attention"
	case abs >= 0.30:
		return "warning"
	default:
		return "info"
	}
}

// buildChangeMessage creates a human-readable explanation.
func buildChangeMessage(label string, prior, current, pctChange float64) string {
	direction := "increased"
	if pctChange < 0 {
		direction = "decreased"
	}
	pctAbs := math.Abs(pctChange) * 100
	return label + " " + direction + " by " +
		formatPct(pctAbs) + "%" +
		" (from " + formatMoney(prior) + " to " + formatMoney(current) + ")"
}

func formatPct(v float64) string {
	s := math.Round(v*10) / 10
	if s == math.Trunc(s) {
		return fmt.Sprintf("%.0f", s)
	}
	return fmt.Sprintf("%.1f", s)
}

func formatMoney(v float64) string {
	return fmt.Sprintf("$%.2f", v)
}
