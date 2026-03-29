package pdf

import "taxpilot/internal/forms"

// Federal1040Mappings returns the PDF field mappings for Form 1040.
// Field IDs extracted from actual 2025 IRS Form 1040 PDF via pdfcpu.
// IDs use XFA-style dot-separated numeric format (e.g., "678.677.840").
// Page 1 fields have prefix "678.677.", page 2 fields have prefix "678.679.".
func Federal1040Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormF1040,
		FormName:     "Form 1040 - U.S. Individual Income Tax Return",
		TemplatePath: "internal/pdf/templates/federal/2025/f1040.pdf",
		Mappings: []FieldMapping{
			// --- Identification (Page 1 top) ---
			{FieldKey: "1040:first_name", PDFField: "678.677.840", Format: "string"},
			{FieldKey: "1040:last_name", PDFField: "678.677.841", Format: "string"},
			{FieldKey: "1040:ssn", PDFField: "678.677.842", Format: "ssn"},

			// Filing status checkboxes (Page 1, fields 832-836)
			{FieldKey: "1040:filing_status_single", PDFField: "678.677.832", Format: "checkbox"},
			{FieldKey: "1040:filing_status_mfj", PDFField: "678.677.833", Format: "checkbox"},
			{FieldKey: "1040:filing_status_mfs", PDFField: "678.677.834", Format: "checkbox"},
			{FieldKey: "1040:filing_status_hoh", PDFField: "678.677.835", Format: "checkbox"},
			{FieldKey: "1040:filing_status_qss", PDFField: "678.677.836", Format: "checkbox"},

			// Spouse info
			{FieldKey: "1040:spouse_name", PDFField: "678.677.843", Format: "string"},
			{FieldKey: "1040:spouse_ssn", PDFField: "678.677.844", Format: "ssn"},

			// Address
			{FieldKey: "1040:address", PDFField: "678.677.848", Format: "string"},
			{FieldKey: "1040:apt", PDFField: "678.677.849", Format: "string"},
			{FieldKey: "1040:city", PDFField: "678.677.850", Format: "string"},
			{FieldKey: "1040:state", PDFField: "678.677.851", Format: "string"},
			{FieldKey: "1040:zip", PDFField: "678.677.852", Format: "string"},
			{FieldKey: "1040:foreign_country", PDFField: "678.677.853", Format: "string"},
			{FieldKey: "1040:foreign_province", PDFField: "678.677.854", Format: "string"},
			{FieldKey: "1040:foreign_postal", PDFField: "678.677.855", Format: "string"},

			// Digital assets question (yes/no checkboxes)
			{FieldKey: "1040:digital_assets_yes", PDFField: "678.677.864", Format: "checkbox"},
			{FieldKey: "1040:digital_assets_no", PDFField: "678.677.865", Format: "checkbox"},

			// --- Income (Page 1 bottom) ---
			// Lines 1a through 1z (wages and related)
			{FieldKey: "1040:1a", PDFField: "678.677.867", Format: "currency"},
			{FieldKey: "1040:1b", PDFField: "678.677.868", Format: "currency"},
			{FieldKey: "1040:1c", PDFField: "678.677.869", Format: "currency"},
			{FieldKey: "1040:1d", PDFField: "678.677.870", Format: "currency"},
			{FieldKey: "1040:1e", PDFField: "678.677.871", Format: "currency"},
			{FieldKey: "1040:1f", PDFField: "678.677.872", Format: "currency"},
			{FieldKey: "1040:1g", PDFField: "678.677.873", Format: "currency"},
			{FieldKey: "1040:1h", PDFField: "678.677.875", Format: "currency"},
			{FieldKey: "1040:1i", PDFField: "678.677.876.902", Format: "currency"},
			{FieldKey: "1040:1z", PDFField: "678.677.879.901", Format: "currency"},

			// Lines 2-9 (interest, dividends, retirement, SS, gains, other)
			{FieldKey: "1040:2a", PDFField: "678.677.880", Format: "currency"},
			{FieldKey: "1040:2b", PDFField: "678.677.881", Format: "currency"},
			{FieldKey: "1040:3a", PDFField: "678.677.882", Format: "currency"},
			{FieldKey: "1040:3b", PDFField: "678.677.885", Format: "currency"},
			{FieldKey: "1040:4a", PDFField: "678.677.886", Format: "currency"},
			{FieldKey: "1040:4b", PDFField: "678.677.887", Format: "currency"},
			{FieldKey: "1040:5a", PDFField: "678.677.888", Format: "currency"},
			{FieldKey: "1040:5b", PDFField: "678.677.889", Format: "currency"},
			{FieldKey: "1040:6a", PDFField: "678.677.890", Format: "currency"},
			{FieldKey: "1040:6b", PDFField: "678.677.891", Format: "currency"},
			{FieldKey: "1040:7", PDFField: "678.677.893", Format: "currency"},
			{FieldKey: "1040:8", PDFField: "678.677.894", Format: "currency"},
			{FieldKey: "1040:9", PDFField: "678.677.895", Format: "currency"},

			// AGI (bottom of Page 1)
			{FieldKey: "1040:10", PDFField: "678.677.897", Format: "currency"},
			{FieldKey: "1040:11", PDFField: "678.677.898", Format: "currency"},

			// --- Deductions (Page 2 top) ---
			// Page 2 fields use prefix "678.679."
			// 680 = name header, 681 = SSN header on page 2
			{FieldKey: "1040:12", PDFField: "678.679.685", Format: "currency"},
			{FieldKey: "1040:13", PDFField: "678.679.687", Format: "currency"},
			{FieldKey: "1040:14", PDFField: "678.679.688", Format: "currency"},
			{FieldKey: "1040:15", PDFField: "678.679.689", Format: "currency"},

			// --- Tax computation (Page 2) ---
			{FieldKey: "1040:16", PDFField: "678.679.690", Format: "currency"},
			{FieldKey: "1040:17", PDFField: "678.679.691", Format: "currency"},
			{FieldKey: "1040:18", PDFField: "678.679.692", Format: "currency"},
			{FieldKey: "1040:19", PDFField: "678.679.694", Format: "currency"},
			{FieldKey: "1040:20", PDFField: "678.679.695", Format: "currency"},
			{FieldKey: "1040:21", PDFField: "678.679.696", Format: "currency"},
			{FieldKey: "1040:22", PDFField: "678.679.697", Format: "currency"},
			{FieldKey: "1040:23", PDFField: "678.679.698", Format: "currency"},
			{FieldKey: "1040:24", PDFField: "678.679.699", Format: "currency"},

			// --- Payments (Page 2) ---
			{FieldKey: "1040:25a", PDFField: "678.679.700", Format: "currency"},
			{FieldKey: "1040:25b", PDFField: "678.679.701", Format: "currency"},
			{FieldKey: "1040:25c", PDFField: "678.679.702", Format: "currency"},
			{FieldKey: "1040:25d", PDFField: "678.679.718", Format: "currency"},
			{FieldKey: "1040:26", PDFField: "678.679.719", Format: "currency"},
			{FieldKey: "1040:27", PDFField: "678.679.720", Format: "currency"},
			{FieldKey: "1040:28", PDFField: "678.679.721", Format: "currency"},
			{FieldKey: "1040:29", PDFField: "678.679.722", Format: "currency"},
			{FieldKey: "1040:30", PDFField: "678.679.723", Format: "currency"},
			{FieldKey: "1040:31", PDFField: "678.679.724", Format: "currency"},
			{FieldKey: "1040:32", PDFField: "678.679.725", Format: "currency"},
			{FieldKey: "1040:33", PDFField: "678.679.726", Format: "currency"},

			// --- Refund / Amount owed (Page 2) ---
			{FieldKey: "1040:34", PDFField: "678.679.727", Format: "currency"},
			{FieldKey: "1040:35a", PDFField: "678.679.728", Format: "currency"},
			{FieldKey: "1040:36", PDFField: "678.679.731", Format: "currency"},
			{FieldKey: "1040:37", PDFField: "678.679.732", Format: "currency"},
			{FieldKey: "1040:38", PDFField: "678.679.735", Format: "currency"},

			// Signature section
			{FieldKey: "1040:occupation", PDFField: "678.679.754", Format: "string"},
		},
	}
}

// ScheduleAMappings returns the PDF field mappings for Schedule A.
func ScheduleAMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormScheduleA,
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
		FormID:       forms.FormScheduleB,
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
		FormID:       forms.FormScheduleC,
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
		FormID:       forms.FormScheduleSE,
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
// Field IDs extracted from actual 2025 IRS Schedule 1 PDF via pdfcpu.
// Page 1 fields have prefix "363.362.", page 2 fields have prefix "363.364.".
func Schedule1Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormSchedule1,
		FormName:     "Schedule 1 — Additional Income and Adjustments to Income",
		TemplatePath: "internal/pdf/templates/federal/2025/schedule_1.pdf",
		Mappings: []FieldMapping{
			// Identification
			{FieldKey: "schedule_1:name", PDFField: "363.362.361", Format: "string"},
			{FieldKey: "schedule_1:ssn", PDFField: "363.362.409", Format: "ssn"},

			// Part I: Additional Income (Page 1)
			{FieldKey: "schedule_1:1", PDFField: "363.362.410", Format: "currency"},      // line 1: taxable refunds
			{FieldKey: "schedule_1:2a", PDFField: "363.362.411", Format: "currency"},     // line 2a: alimony received
			{FieldKey: "schedule_1:3", PDFField: "363.362.414", Format: "currency"},      // line 3: business income (Sch C)
			{FieldKey: "schedule_1:4", PDFField: "363.362.415", Format: "currency"},      // line 4: other gains/losses
			{FieldKey: "schedule_1:5", PDFField: "363.362.416", Format: "currency"},      // line 5: rental real estate
			{FieldKey: "schedule_1:6", PDFField: "363.362.417", Format: "currency"},      // line 6: farm income
			{FieldKey: "schedule_1:7", PDFField: "363.362.418", Format: "currency"},      // line 7: unemployment
			{FieldKey: "schedule_1:8a", PDFField: "363.362.419.442", Format: "currency"}, // line 8a: net operating loss
			{FieldKey: "schedule_1:8b", PDFField: "363.362.420", Format: "currency"},     // line 8b: gambling income
			{FieldKey: "schedule_1:8d", PDFField: "363.362.423", Format: "currency"},     // line 8d: FEIE exclusion (negative)
			{FieldKey: "schedule_1:8z", PDFField: "363.362.435", Format: "currency"},     // line 8z: other income
			{FieldKey: "schedule_1:9", PDFField: "363.362.436", Format: "currency"},      // line 9: total other income
			{FieldKey: "schedule_1:10", PDFField: "363.362.438", Format: "currency"},     // line 10: total additional income

			// Part II: Adjustments to Income (Page 2)
			{FieldKey: "schedule_1:11", PDFField: "363.364.367", Format: "currency"},  // line 11: educator expenses
			{FieldKey: "schedule_1:12", PDFField: "363.364.368", Format: "currency"},  // line 12: business expenses
			{FieldKey: "schedule_1:13", PDFField: "363.364.369", Format: "currency"},  // line 13: HSA deduction
			{FieldKey: "schedule_1:14", PDFField: "363.364.370", Format: "currency"},  // line 14: moving expenses
			{FieldKey: "schedule_1:15", PDFField: "363.364.371", Format: "currency"},  // line 15: deductible SE tax
			{FieldKey: "schedule_1:16", PDFField: "363.364.374", Format: "currency"},  // line 16: SE health insurance
			{FieldKey: "schedule_1:17", PDFField: "363.364.375", Format: "currency"},  // line 17: early withdrawal penalty
			{FieldKey: "schedule_1:18a", PDFField: "363.364.376", Format: "currency"}, // line 18a: IRA deduction
			{FieldKey: "schedule_1:19", PDFField: "363.364.378", Format: "currency"},  // line 19: student loan interest
			{FieldKey: "schedule_1:24z", PDFField: "363.364.395", Format: "currency"}, // line 24z: other adjustments total
			{FieldKey: "schedule_1:25", PDFField: "363.364.400", Format: "currency"},  // line 25: total Part II adjustments
			{FieldKey: "schedule_1:26", PDFField: "363.364.404", Format: "currency"},  // line 26: total adjustments to income
		},
	}
}

// ScheduleDMappings returns the PDF field mappings for Schedule D.
func ScheduleDMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormScheduleD,
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
		FormID:       forms.FormF8949,
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
		FormID:       forms.FormSchedule2,
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
		FormID:       forms.FormSchedule3,
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
		FormID:       forms.FormF8889,
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
		FormID:       forms.FormF8995,
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

// Form2555Mappings returns the PDF field mappings for 2025 Form 2555 (FEIE).
// Field IDs extracted from actual 2025 IRS Form 2555 PDF via pdfcpu.
// Page 1 fields have prefix "433.432.", page 2 "433.434.", page 3 "433.435.".
func Form2555Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormF2555,
		FormName:     "Form 2555 — Foreign Earned Income",
		TemplatePath: "internal/pdf/templates/federal/2025/f2555.pdf",
		Mappings: []FieldMapping{
			// --- Part I: General Information (Page 1) ---
			{FieldKey: "form_2555:name", PDFField: "433.432.431", Format: "string"},
			{FieldKey: "form_2555:ssn", PDFField: "433.432.582.611", Format: "ssn"},
			{FieldKey: "form_2555:foreign_address", PDFField: "433.432.585.610", Format: "string"},
			{FieldKey: "form_2555:occupation", PDFField: "433.432.587", Format: "string"},
			{FieldKey: "form_2555:employer_name_2555", PDFField: "433.432.588", Format: "string"},
			{FieldKey: "form_2555:employer_address", PDFField: "433.432.589", Format: "string"},
			{FieldKey: "form_2555:employer_ein", PDFField: "433.432.591", Format: "string"},
			{FieldKey: "form_2555:employer_foreign_yes", PDFField: "433.432.583", Format: "checkbox"},

			// Part I continued: foreign country and citizenship
			{FieldKey: "form_2555:foreign_country", PDFField: "433.432.586", Format: "string"},
			{FieldKey: "form_2555:citizenship_country", PDFField: "433.432.594", Format: "string"},
			{FieldKey: "form_2555:prior_year_2555", PDFField: "433.432.592", Format: "string"},
			{FieldKey: "form_2555:countries_dates", PDFField: "433.432.596", Format: "string"},

			// --- Part II: Qualifying Tests (Page 2) ---
			{FieldKey: "form_2555:qualifying_test_bfr", PDFField: "433.434.501", Format: "checkbox"},
			{FieldKey: "form_2555:qualifying_test_ppt", PDFField: "433.434.502", Format: "checkbox"},
			{FieldKey: "form_2555:claimed_prior_year", PDFField: "433.434.505", Format: "checkbox"},

			// Physical presence test dates
			{FieldKey: "form_2555:ppt_period_start", PDFField: "433.434.506", Format: "string"},
			{FieldKey: "form_2555:ppt_period_end", PDFField: "433.434.507", Format: "string"},

			// Part III: additional travel info
			{FieldKey: "form_2555:ppt_reason", PDFField: "433.434.511", Format: "string"},
			{FieldKey: "form_2555:ppt_duration", PDFField: "433.434.512", Format: "string"},

			// --- Part IV: Foreign Earned Income (Page 3) ---
			// Page 3 header
			{FieldKey: "form_2555:page3_name", PDFField: "433.435.436", Format: "string"},
			{FieldKey: "form_2555:page3_ssn", PDFField: "433.435.437", Format: "ssn"},

			// Part IV lines
			{FieldKey: "form_2555:foreign_earned_income", PDFField: "433.435.440", Format: "currency"}, // line 24
			{FieldKey: "form_2555:25", PDFField: "433.435.441", Format: "currency"},                    // line 25
			{FieldKey: "form_2555:26", PDFField: "433.435.442", Format: "currency"},                    // line 26
			{FieldKey: "form_2555:27", PDFField: "433.435.443", Format: "currency"},                    // line 27: total
			{FieldKey: "form_2555:27a", PDFField: "433.435.444", Format: "currency"},                   // line 27a
			{FieldKey: "form_2555:28", PDFField: "433.435.445", Format: "currency"},                    // line 28

			// --- Part V/VI: Housing Amount (Page 3) ---
			{FieldKey: "form_2555:29", PDFField: "433.435.446", Format: "currency"}, // line 29
			{FieldKey: "form_2555:30", PDFField: "433.435.447", Format: "currency"}, // line 30
			{FieldKey: "form_2555:31", PDFField: "433.435.448", Format: "currency"}, // line 31
			{FieldKey: "form_2555:32", PDFField: "433.435.449", Format: "currency"}, // line 32
			{FieldKey: "form_2555:33", PDFField: "433.435.450", Format: "currency"}, // line 33
			{FieldKey: "form_2555:34", PDFField: "433.435.451", Format: "currency"}, // line 34
			{FieldKey: "form_2555:35", PDFField: "433.435.452", Format: "currency"}, // line 35
			{FieldKey: "form_2555:36", PDFField: "433.435.453", Format: "currency"}, // line 36

			// --- Part VIII/IX: Exclusion Computation ---
			{FieldKey: "form_2555:exclusion_limit", PDFField: "433.435.456", Format: "currency"},          // line 42: max exclusion
			{FieldKey: "form_2555:ppt_days_present", PDFField: "433.435.457", Format: "integer"},          // line 43: qualifying days
			{FieldKey: "form_2555:qualifying_years", PDFField: "433.435.458", Format: "string"},           // line 44: years fraction
			{FieldKey: "form_2555:prorated_exclusion", PDFField: "433.435.459", Format: "currency"},       // line 45: prorated exclusion
			{FieldKey: "form_2555:foreign_income_exclusion", PDFField: "433.435.460", Format: "currency"}, // line 46: actual exclusion
			{FieldKey: "form_2555:housing_exclusion", PDFField: "433.435.461", Format: "currency"},        // line 47: housing exclusion
			{FieldKey: "form_2555:housing_deduction", PDFField: "433.435.462", Format: "currency"},        // line 48: housing deduction
			{FieldKey: "form_2555:total_exclusion", PDFField: "433.435.465", Format: "currency"},          // line 50: total exclusion
		},
	}
}

// ScheduleCAMappings returns the PDF field mappings for Schedule CA (540).
func ScheduleCAMappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormScheduleCA,
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
		FormID:       forms.FormCA540,
		FormName:     "Form 540 - California Resident Income Tax Return",
		TemplatePath: "internal/pdf/templates/state/ca/2025/f540.pdf",
		Mappings: []FieldMapping{
			// Identification
			{FieldKey: "ca_540:filing_status", PDFField: "540_form_1036 RB", Format: "checkbox"},
			{FieldKey: "1040:first_name", PDFField: "540_form_1003", Format: "string"},
			{FieldKey: "1040:last_name", PDFField: "540_form_1005", Format: "string"},
			{FieldKey: "1040:ssn", PDFField: "540_form_1007", Format: "ssn"},

			// Income (Line 12-19)
			{FieldKey: "ca_540:7", PDFField: "540_form_2018", Format: "currency"},  // Line 12: state wages
			{FieldKey: "ca_540:13", PDFField: "540_form_2019", Format: "currency"}, // Line 13: federal AGI
			{FieldKey: "ca_540:14", PDFField: "540_form_2020", Format: "currency"}, // Line 14: CA subtractions
			{FieldKey: "ca_540:15", PDFField: "540_form_2021", Format: "currency"}, // Line 15: subtract
			{FieldKey: "ca_540:17", PDFField: "540_form_2023", Format: "currency"}, // Line 17: CA AGI
			{FieldKey: "ca_540:18", PDFField: "540_form_2024", Format: "currency"}, // Line 18: deductions
			{FieldKey: "ca_540:19", PDFField: "540_form_2025", Format: "currency"}, // Line 19: taxable income

			// Tax (Line 31-35)
			{FieldKey: "ca_540:31", PDFField: "540_form_2030", Format: "currency"}, // Line 31: tax amount
			{FieldKey: "ca_540:32", PDFField: "540_form_2031", Format: "currency"}, // Line 32: exemption credits
			{FieldKey: "ca_540:35", PDFField: "540_form_2036", Format: "currency"}, // Line 35: subtotal

			// Other Taxes (Line 61-64)
			{FieldKey: "ca_540:36", PDFField: "540_form_3008", Format: "currency"}, // Line 62: Behavioral Health Services Tax
			{FieldKey: "ca_540:40", PDFField: "540_form_3010", Format: "currency"}, // Line 64: total tax

			// Payments (Line 71-78)
			{FieldKey: "ca_540:71", PDFField: "540_form_3011", Format: "currency"}, // Line 71: CA withheld
			{FieldKey: "ca_540:74", PDFField: "540_form_3018", Format: "currency"}, // Line 78: total payments

			// Refund / Amount owed
			{FieldKey: "ca_540:91", PDFField: "540_form_3027", Format: "currency"}, // Line 97: overpaid
			{FieldKey: "ca_540:93", PDFField: "540_form_4005", Format: "currency"}, // Line 100: tax due
		},
	}
}

// Form3514Mappings returns the PDF field mappings for CA Form 3514 (CalEITC).
func Form3514Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       forms.FormF3514,
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
		FormID:       forms.FormF3853,
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
