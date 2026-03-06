package pdf

// Federal1040Mappings returns the PDF field mappings for Form 1040.
// AcroForm field names are reasonable guesses based on common IRS PDF structure.
// These will be updated once actual 2025 PDF templates are available.
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
			{FieldKey: "1040:25d", PDFField: "topmostSubform[0].Page2[0].f2_17[0]", Format: "currency"},
			{FieldKey: "1040:33", PDFField: "topmostSubform[0].Page2[0].f2_25[0]", Format: "currency"},

			// Refund / Amount owed
			{FieldKey: "1040:34", PDFField: "topmostSubform[0].Page2[0].f2_26[0]", Format: "currency"},
			{FieldKey: "1040:37", PDFField: "topmostSubform[0].Page2[0].f2_29[0]", Format: "currency"},
		},
	}
}

// CA540Mappings returns the PDF field mappings for CA Form 540.
// AcroForm field names are reasonable guesses based on common FTB PDF structure.
// These will be updated once actual 2025 PDF templates are available.
func CA540Mappings() *FormPDFConfig {
	return &FormPDFConfig{
		FormID:       "ca_540",
		FormName:     "Form 540 - California Resident Income Tax Return",
		TemplatePath: "internal/pdf/templates/state/ca/2025/f540.pdf",
		Mappings: []FieldMapping{
			// Filing status
			{FieldKey: "ca_540:filing_status", PDFField: "Filing_Status", Format: "checkbox"},

			// Income
			{FieldKey: "ca_540:7", PDFField: "Line_7", Format: "currency"},
			{FieldKey: "ca_540:13", PDFField: "Line_13", Format: "currency"},
			{FieldKey: "ca_540:14", PDFField: "Line_14", Format: "currency"},
			{FieldKey: "ca_540:15", PDFField: "Line_15", Format: "currency"},
			{FieldKey: "ca_540:17", PDFField: "Line_17", Format: "currency"},

			// Deductions
			{FieldKey: "ca_540:18", PDFField: "Line_18", Format: "currency"},
			{FieldKey: "ca_540:19", PDFField: "Line_19", Format: "currency"},

			// Tax
			{FieldKey: "ca_540:31", PDFField: "Line_31", Format: "currency"},
			{FieldKey: "ca_540:32", PDFField: "Line_32", Format: "currency"},
			{FieldKey: "ca_540:35", PDFField: "Line_35", Format: "currency"},
			{FieldKey: "ca_540:36", PDFField: "Line_36", Format: "currency"},
			{FieldKey: "ca_540:40", PDFField: "Line_40", Format: "currency"},

			// Payments
			{FieldKey: "ca_540:71", PDFField: "Line_71", Format: "currency"},
			{FieldKey: "ca_540:74", PDFField: "Line_74", Format: "currency"},

			// Refund / Amount owed
			{FieldKey: "ca_540:91", PDFField: "Line_91", Format: "currency"},
			{FieldKey: "ca_540:93", PDFField: "Line_93", Format: "currency"},
		},
	}
}
