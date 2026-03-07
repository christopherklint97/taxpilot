package pdf

// Federal1040Mappings returns the PDF field mappings for Form 1040.
// Field names verified against actual 2025 IRS Form 1040 PDF AcroForm fields.
func Federal1040Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "1040",
		FormName:     "Form 1040 - U.S. Individual Income Tax Return",
		TemplatePath: "internal/pdf/templates/federal/2025/f1040.pdf",
		Mappings: []FieldMapping{
			// Identification
			{FieldKey: "1040:first_name", PDFField: "topmostSubform[0].Page1[0].f1_02[0]", Format: "string"},
			{FieldKey: "1040:last_name", PDFField: "topmostSubform[0].Page1[0].f1_03[0]", Format: "string"},
			{FieldKey: "1040:ssn", PDFField: "topmostSubform[0].Page1[0].f1_04[0]", Format: "ssn"},
			{FieldKey: "1040:filing_status", PDFField: "topmostSubform[0].Page1[0].c1_1[0]", Format: "checkbox"},

			// Income
			{FieldKey: "1040:1a", PDFField: "topmostSubform[0].Page1[0].f1_07[0]", Format: "currency"},
			{FieldKey: "1040:1z", PDFField: "topmostSubform[0].Page1[0].f1_14[0]", Format: "currency"},
			{FieldKey: "1040:2a", PDFField: "topmostSubform[0].Page1[0].f1_15[0]", Format: "currency"},
			{FieldKey: "1040:2b", PDFField: "topmostSubform[0].Page1[0].f1_16[0]", Format: "currency"},
			{FieldKey: "1040:3a", PDFField: "topmostSubform[0].Page1[0].f1_17[0]", Format: "currency"},
			{FieldKey: "1040:3b", PDFField: "topmostSubform[0].Page1[0].f1_18[0]", Format: "currency"},
			{FieldKey: "1040:7", PDFField: "topmostSubform[0].Page1[0].f1_20[0]", Format: "currency"},
			{FieldKey: "1040:8", PDFField: "topmostSubform[0].Page1[0].f1_21[0]", Format: "currency"},
			{FieldKey: "1040:9", PDFField: "topmostSubform[0].Page1[0].f1_22[0]", Format: "currency"},

			// AGI
			{FieldKey: "1040:10", PDFField: "topmostSubform[0].Page1[0].f1_23[0]", Format: "currency"},
			{FieldKey: "1040:11", PDFField: "topmostSubform[0].Page1[0].f1_24[0]", Format: "currency"},

			// Deductions
			{FieldKey: "1040:12", PDFField: "topmostSubform[0].Page2[0].f2_01[0]", Format: "currency"},
			{FieldKey: "1040:13", PDFField: "topmostSubform[0].Page2[0].f2_02[0]", Format: "currency"},
			{FieldKey: "1040:14", PDFField: "topmostSubform[0].Page2[0].f2_03[0]", Format: "currency"},
			{FieldKey: "1040:15", PDFField: "topmostSubform[0].Page2[0].f2_04[0]", Format: "currency"},

			// Tax
			{FieldKey: "1040:16", PDFField: "topmostSubform[0].Page2[0].f2_05[0]", Format: "currency"},
			{FieldKey: "1040:24", PDFField: "topmostSubform[0].Page2[0].f2_13[0]", Format: "currency"},

			// Payments
			{FieldKey: "1040:25a", PDFField: "topmostSubform[0].Page2[0].f2_14[0]", Format: "currency"},
			{FieldKey: "1040:25b", PDFField: "topmostSubform[0].Page2[0].f2_15[0]", Format: "currency"},
			{FieldKey: "1040:25d", PDFField: "topmostSubform[0].Page2[0].f2_17[0]", Format: "currency"},
			{FieldKey: "1040:33", PDFField: "topmostSubform[0].Page2[0].f2_25[0]", Format: "currency"},

			// Refund / Amount owed
			{FieldKey: "1040:34", PDFField: "topmostSubform[0].Page2[0].f2_26[0]", Format: "currency"},
			{FieldKey: "1040:37", PDFField: "topmostSubform[0].Page2[0].f2_29[0]", Format: "currency"},
		},
	}
}

// ScheduleAMappings returns the PDF field mappings for Schedule A.
func ScheduleAMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_a",
		FormName:     "Schedule A — Itemized Deductions",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_a.pdf",
		Mappings: []FieldMapping{
			// Medical
			{FieldKey: "schedule_a:1", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "currency"},
			{FieldKey: "schedule_a:4", PDFField: "topmostSubform[0].Page1[0].f1_04[0]", Format: "currency"},
			// Taxes
			{FieldKey: "schedule_a:5a", PDFField: "topmostSubform[0].Page1[0].f1_05a[0]", Format: "currency"},
			{FieldKey: "schedule_a:5b", PDFField: "topmostSubform[0].Page1[0].f1_05b[0]", Format: "currency"},
			{FieldKey: "schedule_a:5c", PDFField: "topmostSubform[0].Page1[0].f1_05c[0]", Format: "currency"},
			{FieldKey: "schedule_a:5d", PDFField: "topmostSubform[0].Page1[0].f1_05d[0]", Format: "currency"},
			{FieldKey: "schedule_a:5e", PDFField: "topmostSubform[0].Page1[0].f1_05e[0]", Format: "currency"},
			// Interest
			{FieldKey: "schedule_a:8a", PDFField: "topmostSubform[0].Page1[0].f1_08a[0]", Format: "currency"},
			{FieldKey: "schedule_a:11", PDFField: "topmostSubform[0].Page1[0].f1_11[0]", Format: "currency"},
			// Charitable
			{FieldKey: "schedule_a:12", PDFField: "topmostSubform[0].Page1[0].f1_12[0]", Format: "currency"},
			{FieldKey: "schedule_a:13", PDFField: "topmostSubform[0].Page1[0].f1_13[0]", Format: "currency"},
			{FieldKey: "schedule_a:14", PDFField: "topmostSubform[0].Page1[0].f1_14[0]", Format: "currency"},
			{FieldKey: "schedule_a:15", PDFField: "topmostSubform[0].Page1[0].f1_15[0]", Format: "currency"},
			// Total
			{FieldKey: "schedule_a:17", PDFField: "topmostSubform[0].Page1[0].f1_17[0]", Format: "currency"},
		},
	}
}

// ScheduleBMappings returns the PDF field mappings for Schedule B.
func ScheduleBMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_b",
		FormName:     "Schedule B — Interest and Ordinary Dividends",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_b.pdf",
		Mappings: []FieldMapping{
			// Part I: Interest
			{FieldKey: "schedule_b:1", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "currency"},
			{FieldKey: "schedule_b:4", PDFField: "topmostSubform[0].Page1[0].f1_04[0]", Format: "currency"},
			// Part II: Ordinary Dividends
			{FieldKey: "schedule_b:5", PDFField: "topmostSubform[0].Page1[0].f1_05[0]", Format: "currency"},
			{FieldKey: "schedule_b:6", PDFField: "topmostSubform[0].Page1[0].f1_06[0]", Format: "currency"},
		},
	}
}

// ScheduleCMappings returns the PDF field mappings for Schedule C.
func ScheduleCMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_c",
		FormName:     "Schedule C — Profit or Loss From Business",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_c.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "schedule_c:business_name", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "string"},
			{FieldKey: "schedule_c:business_code", PDFField: "topmostSubform[0].Page1[0].f1_02[0]", Format: "string"},
			{FieldKey: "schedule_c:1", PDFField: "topmostSubform[0].Page1[0].f1_03[0]", Format: "currency"},
			{FieldKey: "schedule_c:5", PDFField: "topmostSubform[0].Page1[0].f1_05[0]", Format: "currency"},
			{FieldKey: "schedule_c:7", PDFField: "topmostSubform[0].Page1[0].f1_07[0]", Format: "currency"},
			{FieldKey: "schedule_c:28", PDFField: "topmostSubform[0].Page1[0].f1_28[0]", Format: "currency"},
			{FieldKey: "schedule_c:31", PDFField: "topmostSubform[0].Page1[0].f1_31[0]", Format: "currency"},
		},
	}
}

// ScheduleSEMappings returns the PDF field mappings for Schedule SE.
func ScheduleSEMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_se",
		FormName:     "Schedule SE — Self-Employment Tax",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_se.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "schedule_se:2", PDFField: "topmostSubform[0].Page1[0].f1_02[0]", Format: "currency"},
			{FieldKey: "schedule_se:3", PDFField: "topmostSubform[0].Page1[0].f1_03[0]", Format: "currency"},
			{FieldKey: "schedule_se:6", PDFField: "topmostSubform[0].Page1[0].f1_06[0]", Format: "currency"},
			{FieldKey: "schedule_se:7", PDFField: "topmostSubform[0].Page1[0].f1_07[0]", Format: "currency"},
		},
	}
}

// Schedule1Mappings returns the PDF field mappings for Schedule 1.
func Schedule1Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_1",
		FormName:     "Schedule 1 — Additional Income and Adjustments to Income",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_1.pdf",
		Mappings: []FieldMapping{
			// Part I: Additional Income
			{FieldKey: "schedule_1:1", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "currency"},
			{FieldKey: "schedule_1:3", PDFField: "topmostSubform[0].Page1[0].f1_03[0]", Format: "currency"},
			{FieldKey: "schedule_1:7", PDFField: "topmostSubform[0].Page1[0].f1_07[0]", Format: "currency"},
			{FieldKey: "schedule_1:10", PDFField: "topmostSubform[0].Page1[0].f1_10[0]", Format: "currency"},
			// Part II: Adjustments
			{FieldKey: "schedule_1:15", PDFField: "topmostSubform[0].Page1[0].f1_15[0]", Format: "currency"},
			{FieldKey: "schedule_1:24", PDFField: "topmostSubform[0].Page1[0].f1_24[0]", Format: "currency"},
			{FieldKey: "schedule_1:26", PDFField: "topmostSubform[0].Page1[0].f1_26[0]", Format: "currency"},
		},
	}
}

// ScheduleDMappings returns the PDF field mappings for Schedule D.
func ScheduleDMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_d",
		FormName:     "Schedule D — Capital Gains and Losses",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_d.pdf",
		Mappings: []FieldMapping{
			// Part I: Short-Term
			{FieldKey: "schedule_d:1", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "currency"},
			{FieldKey: "schedule_d:7", PDFField: "topmostSubform[0].Page1[0].f1_07[0]", Format: "currency"},
			// Part II: Long-Term
			{FieldKey: "schedule_d:8", PDFField: "topmostSubform[0].Page1[0].f1_08[0]", Format: "currency"},
			{FieldKey: "schedule_d:13", PDFField: "topmostSubform[0].Page1[0].f1_13[0]", Format: "currency"},
			{FieldKey: "schedule_d:15", PDFField: "topmostSubform[0].Page1[0].f1_15[0]", Format: "currency"},
			// Part III: Summary
			{FieldKey: "schedule_d:16", PDFField: "topmostSubform[0].Page2[0].f2_16[0]", Format: "currency"},
		},
	}
}

// Form8949Mappings returns the PDF field mappings for Form 8949.
func Form8949Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "form_8949",
		FormName:     "Form 8949 — Sales and Other Dispositions of Capital Assets",
		TemplatePath: "internal/pdf/templates/federal/2025/form_8949.pdf",
		Mappings: []FieldMapping{
			// Part I: Short-Term
			{FieldKey: "form_8949:st_proceeds", PDFField: "topmostSubform[0].Page1[0].f1_st_proceeds[0]", Format: "currency"},
			{FieldKey: "form_8949:st_basis", PDFField: "topmostSubform[0].Page1[0].f1_st_basis[0]", Format: "currency"},
			{FieldKey: "form_8949:st_wash", PDFField: "topmostSubform[0].Page1[0].f1_st_wash[0]", Format: "currency"},
			{FieldKey: "form_8949:st_gain_loss", PDFField: "topmostSubform[0].Page1[0].f1_st_gain[0]", Format: "currency"},
			// Part II: Long-Term
			{FieldKey: "form_8949:lt_proceeds", PDFField: "topmostSubform[0].Page2[0].f2_lt_proceeds[0]", Format: "currency"},
			{FieldKey: "form_8949:lt_basis", PDFField: "topmostSubform[0].Page2[0].f2_lt_basis[0]", Format: "currency"},
			{FieldKey: "form_8949:lt_wash", PDFField: "topmostSubform[0].Page2[0].f2_lt_wash[0]", Format: "currency"},
			{FieldKey: "form_8949:lt_gain_loss", PDFField: "topmostSubform[0].Page2[0].f2_lt_gain[0]", Format: "currency"},
		},
	}
}

// Schedule2Mappings returns the PDF field mappings for Schedule 2.
func Schedule2Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_2",
		FormName:     "Schedule 2 — Additional Taxes",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_2.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "schedule_2:1", PDFField: "topmostSubform[0].Page1[0].f1_01[0]", Format: "currency"},
			{FieldKey: "schedule_2:3", PDFField: "topmostSubform[0].Page1[0].f1_03[0]", Format: "currency"},
			{FieldKey: "schedule_2:6", PDFField: "topmostSubform[0].Page1[0].f1_06[0]", Format: "currency"},
			{FieldKey: "schedule_2:12", PDFField: "topmostSubform[0].Page1[0].f1_12[0]", Format: "currency"},
			{FieldKey: "schedule_2:18", PDFField: "topmostSubform[0].Page1[0].f1_18[0]", Format: "currency"},
			{FieldKey: "schedule_2:21", PDFField: "topmostSubform[0].Page1[0].f1_21[0]", Format: "currency"},
		},
	}
}

// Schedule3Mappings returns the PDF field mappings for Schedule 3.
func Schedule3Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "schedule_3",
		FormName:     "Schedule 3 — Additional Credits and Payments",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_3.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "schedule_3:8", PDFField: "topmostSubform[0].Page1[0].f1_08[0]", Format: "currency"},
			{FieldKey: "schedule_3:10", PDFField: "topmostSubform[0].Page1[0].f1_10[0]", Format: "currency"},
			{FieldKey: "schedule_3:15", PDFField: "topmostSubform[0].Page1[0].f1_15[0]", Format: "currency"},
		},
	}
}

// Form8889Mappings returns the PDF field mappings for Form 8889.
func Form8889Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "form_8889",
		FormName:     "Form 8889 — Health Savings Accounts",
		TemplatePath: "internal/pdf/templates/federal/2025/f8889.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "form_8889:1", PDFField: "f8889_01", Format: "string"},
			{FieldKey: "form_8889:2", PDFField: "f8889_02", Format: "currency"},
			{FieldKey: "form_8889:3", PDFField: "f8889_03", Format: "currency"},
			{FieldKey: "form_8889:5", PDFField: "f8889_05", Format: "currency"},
			{FieldKey: "form_8889:6", PDFField: "f8889_06", Format: "currency"},
			{FieldKey: "form_8889:9", PDFField: "f8889_09", Format: "currency"},
			{FieldKey: "form_8889:14a", PDFField: "f8889_14a", Format: "currency"},
			{FieldKey: "form_8889:14c", PDFField: "f8889_14c", Format: "currency"},
			{FieldKey: "form_8889:15", PDFField: "f8889_15", Format: "currency"},
			{FieldKey: "form_8889:17b", PDFField: "f8889_17b", Format: "currency"},
		},
	}
}

// Form8995Mappings returns the PDF field mappings for Form 8995.
func Form8995Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "form_8995",
		FormName:     "Form 8995 — Qualified Business Income Deduction (Simplified)",
		TemplatePath: "internal/pdf/templates/federal/2025/f8995.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "form_8995:1", PDFField: "f8995_01", Format: "currency"},
			{FieldKey: "form_8995:2", PDFField: "f8995_02", Format: "currency"},
			{FieldKey: "form_8995:3", PDFField: "f8995_03", Format: "currency"},
			{FieldKey: "form_8995:4", PDFField: "f8995_04", Format: "currency"},
			{FieldKey: "form_8995:5", PDFField: "f8995_05", Format: "currency"},
			{FieldKey: "form_8995:6", PDFField: "f8995_06", Format: "currency"},
			{FieldKey: "form_8995:7", PDFField: "f8995_07", Format: "currency"},
			{FieldKey: "form_8995:8", PDFField: "f8995_08", Format: "currency"},
			{FieldKey: "form_8995:10", PDFField: "f8995_10", Format: "currency"},
		},
	}
}

// ScheduleCAMappings returns the PDF field mappings for Schedule CA (540).
func ScheduleCAMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "ca_schedule_ca",
		FormName:     "Schedule CA (540) — California Adjustments",
		TemplatePath: "internal/pdf/templates/state/ca/2025/schedule_ca.pdf",
		Mappings: []FieldMapping{
			// Line 2: Interest
			{FieldKey: "ca_schedule_ca:2_col_a", PDFField: "Line_2_ColA", Format: "currency"},
			{FieldKey: "ca_schedule_ca:2_col_b", PDFField: "Line_2_ColB", Format: "currency"},
			{FieldKey: "ca_schedule_ca:2_col_c", PDFField: "Line_2_ColC", Format: "currency"},
			// Line 3: Dividends
			{FieldKey: "ca_schedule_ca:3_col_a", PDFField: "Line_3_ColA", Format: "currency"},
			{FieldKey: "ca_schedule_ca:3_col_b", PDFField: "Line_3_ColB", Format: "currency"},
			{FieldKey: "ca_schedule_ca:3_col_c", PDFField: "Line_3_ColC", Format: "currency"},
			// Line 7: Capital gains
			{FieldKey: "ca_schedule_ca:7_col_a", PDFField: "Line_7_ColA", Format: "currency"},
			{FieldKey: "ca_schedule_ca:7_col_b", PDFField: "Line_7_ColB", Format: "currency"},
			{FieldKey: "ca_schedule_ca:7_col_c", PDFField: "Line_7_ColC", Format: "currency"},
			// Line 12: Business income
			{FieldKey: "ca_schedule_ca:12_col_b", PDFField: "Line_12_ColB", Format: "currency"},
			{FieldKey: "ca_schedule_ca:12_col_c", PDFField: "Line_12_ColC", Format: "currency"},
			// Line 15: HSA add-back
			{FieldKey: "ca_schedule_ca:15_col_c", PDFField: "Line_15_ColC", Format: "currency"},
			// Line 16: SE tax deduction
			{FieldKey: "ca_schedule_ca:16_col_b", PDFField: "Line_16_ColB", Format: "currency"},
			// Part II: Itemized deduction adjustments
			{FieldKey: "ca_schedule_ca:5a_col_b", PDFField: "Line_5a_ColB_P2", Format: "currency"},
			{FieldKey: "ca_schedule_ca:5e_col_b", PDFField: "Line_5e_ColB_P2", Format: "currency"},
			{FieldKey: "ca_schedule_ca:5e_col_c", PDFField: "Line_5e_ColC_P2", Format: "currency"},
			{FieldKey: "ca_schedule_ca:ca_itemized", PDFField: "CA_Itemized_Total", Format: "currency"},
			// Line 37: Totals
			{FieldKey: "ca_schedule_ca:37_col_a", PDFField: "Line_37_ColA", Format: "currency"},
			{FieldKey: "ca_schedule_ca:37_col_b", PDFField: "Line_37_ColB", Format: "currency"},
			{FieldKey: "ca_schedule_ca:37_col_c", PDFField: "Line_37_ColC", Format: "currency"},
		},
	}
}

// CA540Mappings returns the PDF field mappings for CA Form 540.
// Field names verified against actual 2025 FTB Form 540 PDF AcroForm fields.
func CA540Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "ca_540",
		FormName:     "Form 540 - California Resident Income Tax Return",
		TemplatePath: "internal/pdf/templates/state/ca/2025/f540.pdf",
		Mappings: []FieldMapping{
			// Identification
			{FieldKey: "ca_540:filing_status", PDFField: "540_form_1036 RB", Format: "checkbox"},
			{FieldKey: "1040:first_name", PDFField: "540_form_1003", Format: "string"},
			{FieldKey: "1040:last_name", PDFField: "540_form_1005", Format: "string"},
			{FieldKey: "1040:ssn", PDFField: "540_form_1007", Format: "ssn"},

			// Income (Line 12-19)
			{FieldKey: "ca_540:7", PDFField: "540_form_2018", Format: "currency"},   // Line 12: state wages
			{FieldKey: "ca_540:13", PDFField: "540_form_2019", Format: "currency"},  // Line 13: federal AGI
			{FieldKey: "ca_540:14", PDFField: "540_form_2020", Format: "currency"},  // Line 14: CA subtractions
			{FieldKey: "ca_540:15", PDFField: "540_form_2021", Format: "currency"},  // Line 15: subtract
			{FieldKey: "ca_540:17", PDFField: "540_form_2023", Format: "currency"},  // Line 17: CA AGI
			{FieldKey: "ca_540:18", PDFField: "540_form_2024", Format: "currency"},  // Line 18: deductions
			{FieldKey: "ca_540:19", PDFField: "540_form_2025", Format: "currency"},  // Line 19: taxable income

			// Tax (Line 31-35)
			{FieldKey: "ca_540:31", PDFField: "540_form_2030", Format: "currency"},  // Line 31: tax amount
			{FieldKey: "ca_540:32", PDFField: "540_form_2031", Format: "currency"},  // Line 32: exemption credits
			{FieldKey: "ca_540:35", PDFField: "540_form_2036", Format: "currency"},  // Line 35: subtotal

			// Other Taxes (Line 61-64)
			{FieldKey: "ca_540:36", PDFField: "540_form_3008", Format: "currency"},  // Line 62: Behavioral Health Services Tax
			{FieldKey: "ca_540:40", PDFField: "540_form_3010", Format: "currency"},  // Line 64: total tax

			// Payments (Line 71-78)
			{FieldKey: "ca_540:71", PDFField: "540_form_3011", Format: "currency"},  // Line 71: CA withheld
			{FieldKey: "ca_540:74", PDFField: "540_form_3018", Format: "currency"},  // Line 78: total payments

			// Refund / Amount owed
			{FieldKey: "ca_540:91", PDFField: "540_form_3027", Format: "currency"},  // Line 97: overpaid
			{FieldKey: "ca_540:93", PDFField: "540_form_4005", Format: "currency"},  // Line 100: tax due
		},
	}
}

// Form3514Mappings returns the PDF field mappings for CA Form 3514 (CalEITC).
func Form3514Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "form_3514",
		FormName:     "Form 3514 — California Earned Income Tax Credit",
		TemplatePath: "internal/pdf/templates/state/ca/2025/form_3514.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "form_3514:1", PDFField: "F3514_Line1", Format: "currency"},
			{FieldKey: "form_3514:2", PDFField: "F3514_Line2", Format: "currency"},
			{FieldKey: "form_3514:3", PDFField: "F3514_Line3", Format: "currency"},
			{FieldKey: "form_3514:5", PDFField: "F3514_Line5", Format: "currency"},
			{FieldKey: "form_3514:6", PDFField: "F3514_Line6", Format: "currency"},
			{FieldKey: "form_3514:7", PDFField: "F3514_Line7", Format: "currency"},
		},
	}
}

// Form3853Mappings returns the PDF field mappings for CA Form 3853 (Health Coverage).
func Form3853Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "form_3853",
		FormName:     "Form 3853 — Health Coverage Exemptions and Individual Shared Responsibility Penalty",
		TemplatePath: "internal/pdf/templates/state/ca/2025/form_3853.pdf",
		Mappings: []FieldMapping{
			{FieldKey: "form_3853:1", PDFField: "F3853_Line1", Format: "string"},
			{FieldKey: "form_3853:2", PDFField: "F3853_Line2", Format: "currency"},
			{FieldKey: "form_3853:4", PDFField: "F3853_Line4", Format: "currency"},
			{FieldKey: "form_3853:5", PDFField: "F3853_Line5", Format: "currency"},
			{FieldKey: "form_3853:6", PDFField: "F3853_Line6", Format: "currency"},
			{FieldKey: "form_3853:7", PDFField: "F3853_Line7", Format: "currency"},
		},
	}
}
