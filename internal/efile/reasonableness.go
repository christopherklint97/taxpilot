package efile

import (
	"fmt"

	"taxpilot/internal/forms"
	"math"
)

// ReasonablenessCheck validates computed tax return values for internal
// consistency and flags unusual values that may indicate errors or audit
// triggers. Unlike ValidateReturn, these checks never produce SeverityError
// results — they are advisory warnings and informational notices.
func ReasonablenessCheck(results map[string]float64, strInputs map[string]string, taxYear int, includeCA bool) ValidationReport {
	var report ValidationReport

	filingStatus := getStr(strInputs, forms.F1040FilingStatus)

	// --- Cross-check computed values ---

	// RC001: AGI should equal total income minus adjustments
	if numExists(results, forms.F1040Line11) && numExists(results, forms.F1040Line9) {
		agi := getNum(results, forms.F1040Line11)
		totalIncome := getNum(results, forms.F1040Line9)
		adjustments := getNum(results, forms.Sched1Line26) // 0 if not present
		expected := totalIncome - adjustments
		if math.Abs(agi-expected) > 1.0 {
			report.addResult("RC001", SeverityWarning, forms.F1040Line11,
				fmt.Sprintf("AGI ($%.0f) does not equal total income ($%.0f) minus adjustments ($%.0f)", agi, totalIncome, adjustments))
		}
	}

	// RC002: Taxable income should equal AGI minus deduction
	if numExists(results, forms.F1040Line15) && numExists(results, forms.F1040Line11) {
		taxableIncome := getNum(results, forms.F1040Line15)
		agi := getNum(results, forms.F1040Line11)
		// Use itemized deduction if present, otherwise standard deduction
		deduction := getNum(results, forms.F1040Line12)
		if numExists(results, forms.SchedALine17) && getNum(results, forms.SchedALine17) > 0 {
			deduction = getNum(results, forms.SchedALine17)
		}
		expected := agi - deduction
		if expected < 0 {
			expected = 0
		}
		if math.Abs(taxableIncome-expected) > 1.0 {
			report.addResult("RC002", SeverityWarning, forms.F1040Line15,
				fmt.Sprintf("Taxable income ($%.0f) does not equal AGI ($%.0f) minus deduction ($%.0f)", taxableIncome, agi, deduction))
		}
	}

	// RC003: Refund should equal payments minus tax when payments > tax
	if numExists(results, forms.F1040Line34) && numExists(results, forms.F1040Line33) && numExists(results, forms.F1040Line24) {
		refund := getNum(results, forms.F1040Line34)
		payments := getNum(results, forms.F1040Line33)
		tax := getNum(results, forms.F1040Line24)
		if payments > tax && refund > 0 {
			expected := payments - tax
			if math.Abs(refund-expected) > 1.0 {
				report.addResult("RC003", SeverityWarning, forms.F1040Line34,
					fmt.Sprintf("Refund ($%.0f) does not equal payments ($%.0f) minus tax ($%.0f)", refund, payments, tax))
			}
		}
	}

	// RC004: Amount owed should equal tax minus payments when tax > payments
	if numExists(results, forms.F1040Line37) && numExists(results, forms.F1040Line33) && numExists(results, forms.F1040Line24) {
		owed := getNum(results, forms.F1040Line37)
		payments := getNum(results, forms.F1040Line33)
		tax := getNum(results, forms.F1040Line24)
		if tax > payments && owed > 0 {
			expected := tax - payments
			if math.Abs(owed-expected) > 1.0 {
				report.addResult("RC004", SeverityWarning, forms.F1040Line37,
					fmt.Sprintf("Amount owed ($%.0f) does not equal tax ($%.0f) minus payments ($%.0f)", owed, tax, payments))
			}
		}
	}

	// --- Flag unusual values (audit triggers) ---

	// RC005: Home office deduction > 30% of business income
	if numExists(results, forms.FK(forms.FormScheduleC, "30")) && numExists(results, forms.SchedCLine7) {
		homeOffice := getNum(results, forms.FK(forms.FormScheduleC, "30"))
		bizIncome := getNum(results, forms.SchedCLine7)
		if bizIncome > 0 && homeOffice > 0.3*bizIncome {
			report.addResult("RC005", SeverityWarning, forms.FK(forms.FormScheduleC, "30"),
				"Home office deduction is large relative to business income")
		}
	}

	// RC006: Self-employment income > $400 but no SE tax
	if numExists(results, forms.SchedCLine31) {
		seIncome := getNum(results, forms.SchedCLine31)
		seTax := getNum(results, forms.FK(forms.FormSchedule2, "4"))
		if seIncome > 400 && seTax == 0 {
			report.addResult("RC006", SeverityWarning, forms.FK(forms.FormSchedule2, "4"),
				"Self-employment income present but no SE tax computed")
		}
	}

	// RC007: HSA contributions exceed IRS limits
	if numExists(results, forms.F8889Line2) {
		hsaContrib := getNum(results, forms.F8889Line2)
		// 2025 limits
		singleLimit := 4300.0
		familyLimit := 8550.0
		if hsaContrib > 0 {
			limit := singleLimit
			label := "single"
			if filingStatus == "mfj" || filingStatus == "mfs" {
				limit = familyLimit
				label = "family"
			}
			if hsaContrib > limit {
				report.addResult("RC007", SeverityWarning, forms.F8889Line2,
					fmt.Sprintf("HSA contributions ($%.0f) exceed %s limit ($%.0f) for %d", hsaContrib, label, limit, taxYear))
			}
		}
	}

	// RC008: Capital losses exceed limit
	if numExists(results, forms.FK(forms.FormScheduleD, "21")) {
		capLoss := getNum(results, forms.FK(forms.FormScheduleD, "21"))
		if capLoss < 0 {
			limit := -3000.0
			if filingStatus == "mfs" {
				limit = -1500.0
			}
			if capLoss < limit {
				report.addResult("RC008", SeverityWarning, forms.FK(forms.FormScheduleD, "21"),
					fmt.Sprintf("Capital losses ($%.0f) exceed the deductible limit ($%.0f)", capLoss, limit))
			}
		}
	}

	// RC009: Estimated tax payments exceed total tax
	if numExists(results, forms.Sched3Line8) && numExists(results, forms.F1040Line24) {
		estimated := getNum(results, forms.Sched3Line8)
		totalTax := getNum(results, forms.F1040Line24)
		if estimated > 0 && estimated > totalTax {
			report.addResult("RC009", SeverityInfo, forms.Sched3Line8,
				"Estimated payments exceed total tax; verify amounts")
		}
	}

	// RC010: Effective federal tax rate > 37%
	if numExists(results, forms.F1040Line24) && numExists(results, forms.F1040Line11) {
		totalTax := getNum(results, forms.F1040Line24)
		agi := getNum(results, forms.F1040Line11)
		if agi > 0 && totalTax/agi > 0.37 {
			report.addResult("RC010", SeverityWarning, forms.F1040Line24,
				"Effective tax rate exceeds highest bracket")
		}
	}

	// --- CA consistency checks ---
	if includeCA {
		// RC011: CA AGI differs significantly from federal AGI
		if numExists(results, forms.CA540Line17) && numExists(results, forms.F1040Line11) {
			caAGI := getNum(results, forms.CA540Line17)
			fedAGI := getNum(results, forms.F1040Line11)
			diff := math.Abs(caAGI - fedAGI)
			if fedAGI > 0 && diff > 0.2*math.Abs(fedAGI) && diff > 5000 {
				report.addResult("RC011", SeverityInfo, forms.CA540Line17,
					fmt.Sprintf("CA AGI ($%.0f) differs from federal AGI ($%.0f) by more than 20%%; review Schedule CA adjustments", caAGI, fedAGI))
			}
		}

		// RC012: CA tax should be <= 13.3% of taxable income + 1% mental health surcharge on amount > $1M
		if numExists(results, forms.CA540Line31) && numExists(results, forms.CA540Line19) {
			caTax := getNum(results, forms.CA540Line31)
			caTaxableIncome := getNum(results, forms.CA540Line19)
			if caTaxableIncome > 0 {
				maxTax := 0.133 * caTaxableIncome
				if caTaxableIncome > 1000000 {
					maxTax += 0.01 * (caTaxableIncome - 1000000)
				}
				if caTax > maxTax {
					report.addResult("RC012", SeverityWarning, forms.CA540Line31,
						fmt.Sprintf("CA tax ($%.0f) exceeds maximum expected rate for taxable income ($%.0f)", caTax, caTaxableIncome))
				}
			}
		}

		// RC013: HSA add-back on Schedule CA
		if numExists(results, forms.F8889Line2) && getNum(results, forms.F8889Line2) > 0 {
			if !numExists(results, forms.FK(forms.FormScheduleCA, "15_col_c")) || getNum(results, forms.FK(forms.FormScheduleCA, "15_col_c")) <= 0 {
				report.addResult("RC013", SeverityInfo, forms.FK(forms.FormScheduleCA, "15_col_c"),
					"HSA contributions taken federally but no CA Schedule CA add-back found; CA does not conform to federal HSA deduction")
			}
		}

		// RC014: QBI deduction should not carry to CA
		if numExists(results, forms.FK(forms.FormF8995, "15")) && getNum(results, forms.FK(forms.FormF8995, "15")) > 0 {
			if numExists(results, forms.CA540Line18) {
				caDeduction := getNum(results, forms.CA540Line18)
				fedDeduction := getNum(results, forms.F1040Line12)
				qbi := getNum(results, forms.FK(forms.FormF8995, "15"))
				// If CA deduction includes QBI, it would be close to federal deduction
				// CA should not include QBI, so CA deduction should be less
				if fedDeduction > 0 && caDeduction >= fedDeduction && qbi > 0 {
					report.addResult("RC014", SeverityInfo, forms.CA540Line18,
						"QBI deduction taken federally; verify CA does not include QBI deduction (CA does not conform)")
				}
			}
		}
	}

	// --- Expat reasonableness checks ---

	// RC015: FEIE claimed but low qualifying days
	if numExists(results, forms.F2555TotalExclusion) && getNum(results, forms.F2555TotalExclusion) > 0 {
		days := getNum(results, forms.F2555QualifyingDays)
		if days > 0 && days < 330 {
			report.addResult("RC015", SeverityWarning, forms.F2555QualifyingDays,
				fmt.Sprintf("FEIE claimed with only %d qualifying days; full exclusion requires 365 days (BFRT) or 330 days (PPT)", int(days)))
		}
	}

	// RC016: FTC exceeds 50% of foreign income
	if numExists(results, forms.F1116Line22) && numExists(results, forms.F1116Line7) {
		ftc := getNum(results, forms.F1116Line22)
		foreignIncome := getNum(results, forms.F1116Line7)
		if foreignIncome > 0 && ftc > 0.5*foreignIncome {
			report.addResult("RC016", SeverityWarning, forms.F1116Line22,
				"Foreign tax credit exceeds 50% of foreign source income; verify foreign taxes paid")
		}
	}

	// RC017: FEIE + FTC no double benefit check
	if numExists(results, forms.F2555TotalExclusion) && numExists(results, forms.F1116Line22) {
		feie := getNum(results, forms.F2555TotalExclusion)
		ftc := getNum(results, forms.F1116Line22)
		if feie > 0 && ftc > 0 {
			report.addResult("RC017", SeverityInfo, forms.F1116Line22,
				"Both FEIE and FTC claimed — verify FTC applies only to income NOT excluded by FEIE")
		}
	}

	// RC018: FATCA value fluctuation (max value >> year-end value)
	if numExists(results, forms.F8938TotalMaxValue) && numExists(results, forms.F8938TotalYearEndValue) {
		maxVal := getNum(results, forms.F8938TotalMaxValue)
		yearEnd := getNum(results, forms.F8938TotalYearEndValue)
		if yearEnd > 0 && maxVal > 3*yearEnd {
			report.addResult("RC018", SeverityInfo, forms.F8938TotalMaxValue,
				fmt.Sprintf("Peak foreign asset value ($%.0f) is more than 3x year-end value ($%.0f); verify values", maxVal, yearEnd))
		}
	}

	if includeCA {
		// RC019: CA FEIE add-back equals federal FEIE
		if numExists(results, forms.F2555TotalExclusion) && getNum(results, forms.F2555TotalExclusion) > 0 {
			feie := getNum(results, forms.F2555TotalExclusion)
			addBack := getNum(results, forms.SchedCALine8dColC)
			if math.Abs(addBack-feie) > 1.0 {
				report.addResult("RC019", SeverityWarning, forms.SchedCALine8dColC,
					fmt.Sprintf("CA FEIE add-back ($%.0f) does not match federal FEIE ($%.0f)", addBack, feie))
			}
		}
	}

	report.computeValidity()
	return report
}
