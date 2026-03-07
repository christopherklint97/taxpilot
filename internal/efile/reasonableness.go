package efile

import (
	"fmt"
	"math"
)

// ReasonablenessCheck validates computed tax return values for internal
// consistency and flags unusual values that may indicate errors or audit
// triggers. Unlike ValidateReturn, these checks never produce SeverityError
// results — they are advisory warnings and informational notices.
func ReasonablenessCheck(results map[string]float64, strInputs map[string]string, taxYear int, includeCA bool) ValidationReport {
	var report ValidationReport

	filingStatus := getStr(strInputs, "1040:filing_status")

	// --- Cross-check computed values ---

	// RC001: AGI should equal total income minus adjustments
	if numExists(results, "1040:11") && numExists(results, "1040:9") {
		agi := getNum(results, "1040:11")
		totalIncome := getNum(results, "1040:9")
		adjustments := getNum(results, "schedule_1:26") // 0 if not present
		expected := totalIncome - adjustments
		if math.Abs(agi-expected) > 1.0 {
			report.addResult("RC001", SeverityWarning, "1040:11",
				fmt.Sprintf("AGI ($%.0f) does not equal total income ($%.0f) minus adjustments ($%.0f)", agi, totalIncome, adjustments))
		}
	}

	// RC002: Taxable income should equal AGI minus deduction
	if numExists(results, "1040:15") && numExists(results, "1040:11") {
		taxableIncome := getNum(results, "1040:15")
		agi := getNum(results, "1040:11")
		// Use itemized deduction if present, otherwise standard deduction
		deduction := getNum(results, "1040:12")
		if numExists(results, "schedule_a:17") && getNum(results, "schedule_a:17") > 0 {
			deduction = getNum(results, "schedule_a:17")
		}
		expected := agi - deduction
		if expected < 0 {
			expected = 0
		}
		if math.Abs(taxableIncome-expected) > 1.0 {
			report.addResult("RC002", SeverityWarning, "1040:15",
				fmt.Sprintf("Taxable income ($%.0f) does not equal AGI ($%.0f) minus deduction ($%.0f)", taxableIncome, agi, deduction))
		}
	}

	// RC003: Refund should equal payments minus tax when payments > tax
	if numExists(results, "1040:34") && numExists(results, "1040:33") && numExists(results, "1040:24") {
		refund := getNum(results, "1040:34")
		payments := getNum(results, "1040:33")
		tax := getNum(results, "1040:24")
		if payments > tax && refund > 0 {
			expected := payments - tax
			if math.Abs(refund-expected) > 1.0 {
				report.addResult("RC003", SeverityWarning, "1040:34",
					fmt.Sprintf("Refund ($%.0f) does not equal payments ($%.0f) minus tax ($%.0f)", refund, payments, tax))
			}
		}
	}

	// RC004: Amount owed should equal tax minus payments when tax > payments
	if numExists(results, "1040:37") && numExists(results, "1040:33") && numExists(results, "1040:24") {
		owed := getNum(results, "1040:37")
		payments := getNum(results, "1040:33")
		tax := getNum(results, "1040:24")
		if tax > payments && owed > 0 {
			expected := tax - payments
			if math.Abs(owed-expected) > 1.0 {
				report.addResult("RC004", SeverityWarning, "1040:37",
					fmt.Sprintf("Amount owed ($%.0f) does not equal tax ($%.0f) minus payments ($%.0f)", owed, tax, payments))
			}
		}
	}

	// --- Flag unusual values (audit triggers) ---

	// RC005: Home office deduction > 30% of business income
	if numExists(results, "schedule_c:30") && numExists(results, "schedule_c:7") {
		homeOffice := getNum(results, "schedule_c:30")
		bizIncome := getNum(results, "schedule_c:7")
		if bizIncome > 0 && homeOffice > 0.3*bizIncome {
			report.addResult("RC005", SeverityWarning, "schedule_c:30",
				"Home office deduction is large relative to business income")
		}
	}

	// RC006: Self-employment income > $400 but no SE tax
	if numExists(results, "schedule_c:31") {
		seIncome := getNum(results, "schedule_c:31")
		seTax := getNum(results, "schedule_2:4")
		if seIncome > 400 && seTax == 0 {
			report.addResult("RC006", SeverityWarning, "schedule_2:4",
				"Self-employment income present but no SE tax computed")
		}
	}

	// RC007: HSA contributions exceed IRS limits
	if numExists(results, "form_8889:2") {
		hsaContrib := getNum(results, "form_8889:2")
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
				report.addResult("RC007", SeverityWarning, "form_8889:2",
					fmt.Sprintf("HSA contributions ($%.0f) exceed %s limit ($%.0f) for %d", hsaContrib, label, limit, taxYear))
			}
		}
	}

	// RC008: Capital losses exceed limit
	if numExists(results, "schedule_d:21") {
		capLoss := getNum(results, "schedule_d:21")
		if capLoss < 0 {
			limit := -3000.0
			if filingStatus == "mfs" {
				limit = -1500.0
			}
			if capLoss < limit {
				report.addResult("RC008", SeverityWarning, "schedule_d:21",
					fmt.Sprintf("Capital losses ($%.0f) exceed the deductible limit ($%.0f)", capLoss, limit))
			}
		}
	}

	// RC009: Estimated tax payments exceed total tax
	if numExists(results, "schedule_3:8") && numExists(results, "1040:24") {
		estimated := getNum(results, "schedule_3:8")
		totalTax := getNum(results, "1040:24")
		if estimated > 0 && estimated > totalTax {
			report.addResult("RC009", SeverityInfo, "schedule_3:8",
				"Estimated payments exceed total tax; verify amounts")
		}
	}

	// RC010: Effective federal tax rate > 37%
	if numExists(results, "1040:24") && numExists(results, "1040:11") {
		totalTax := getNum(results, "1040:24")
		agi := getNum(results, "1040:11")
		if agi > 0 && totalTax/agi > 0.37 {
			report.addResult("RC010", SeverityWarning, "1040:24",
				"Effective tax rate exceeds highest bracket")
		}
	}

	// --- CA consistency checks ---
	if includeCA {
		// RC011: CA AGI differs significantly from federal AGI
		if numExists(results, "ca_540:17") && numExists(results, "1040:11") {
			caAGI := getNum(results, "ca_540:17")
			fedAGI := getNum(results, "1040:11")
			diff := math.Abs(caAGI - fedAGI)
			if fedAGI > 0 && diff > 0.2*math.Abs(fedAGI) && diff > 5000 {
				report.addResult("RC011", SeverityInfo, "ca_540:17",
					fmt.Sprintf("CA AGI ($%.0f) differs from federal AGI ($%.0f) by more than 20%%; review Schedule CA adjustments", caAGI, fedAGI))
			}
		}

		// RC012: CA tax should be <= 13.3% of taxable income + 1% mental health surcharge on amount > $1M
		if numExists(results, "ca_540:31") && numExists(results, "ca_540:19") {
			caTax := getNum(results, "ca_540:31")
			caTaxableIncome := getNum(results, "ca_540:19")
			if caTaxableIncome > 0 {
				maxTax := 0.133 * caTaxableIncome
				if caTaxableIncome > 1000000 {
					maxTax += 0.01 * (caTaxableIncome - 1000000)
				}
				if caTax > maxTax {
					report.addResult("RC012", SeverityWarning, "ca_540:31",
						fmt.Sprintf("CA tax ($%.0f) exceeds maximum expected rate for taxable income ($%.0f)", caTax, caTaxableIncome))
				}
			}
		}

		// RC013: HSA add-back on Schedule CA
		if numExists(results, "form_8889:2") && getNum(results, "form_8889:2") > 0 {
			if !numExists(results, "ca_schedule_ca:15_col_c") || getNum(results, "ca_schedule_ca:15_col_c") <= 0 {
				report.addResult("RC013", SeverityInfo, "ca_schedule_ca:15_col_c",
					"HSA contributions taken federally but no CA Schedule CA add-back found; CA does not conform to federal HSA deduction")
			}
		}

		// RC014: QBI deduction should not carry to CA
		if numExists(results, "form_8995:15") && getNum(results, "form_8995:15") > 0 {
			if numExists(results, "ca_540:18") {
				caDeduction := getNum(results, "ca_540:18")
				fedDeduction := getNum(results, "1040:12")
				qbi := getNum(results, "form_8995:15")
				// If CA deduction includes QBI, it would be close to federal deduction
				// CA should not include QBI, so CA deduction should be less
				if fedDeduction > 0 && caDeduction >= fedDeduction && qbi > 0 {
					report.addResult("RC014", SeverityInfo, "ca_540:18",
						"QBI deduction taken federally; verify CA does not include QBI deduction (CA does not conform)")
				}
			}
		}
	}

	report.computeValidity()
	return report
}
