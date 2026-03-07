package knowledge

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractFromText_LinePatterns(t *testing.T) {
	text := `General Information
This form is used for reporting income.

Line 1. Wages, salaries, tips
Enter the total from all W-2 forms Box 1.
Include all compensation received during the tax year from employers.
This includes tips and other forms of compensation reported on your W-2.

Line 2. Interest income
Enter taxable interest from Form 1099-INT.
Include interest from banks, savings accounts, and bonds.

Line 3. Dividend income
Enter ordinary dividends from Form 1099-DIV.
Qualified dividends are taxed at capital gains rates.
`

	docs := ExtractFromText(text, "Form 1040 Instructions", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) < 3 {
		t.Fatalf("expected at least 3 chunks from Line N patterns, got %d", len(docs))
	}

	// Check that Line 1 chunk exists and has the right content
	foundLine1 := false
	foundLine2 := false
	for _, doc := range docs {
		if strings.Contains(doc.Section, "Line 1") {
			foundLine1 = true
			if !strings.Contains(doc.Content, "W-2") {
				t.Errorf("Line 1 chunk should contain 'W-2', got: %s", doc.Content)
			}
		}
		if strings.Contains(doc.Section, "Line 2") {
			foundLine2 = true
			if !strings.Contains(doc.Content, "1099-INT") {
				t.Errorf("Line 2 chunk should contain '1099-INT', got: %s", doc.Content)
			}
		}
	}
	if !foundLine1 {
		t.Error("expected to find a chunk for Line 1")
	}
	if !foundLine2 {
		t.Error("expected to find a chunk for Line 2")
	}
}

func TestExtractFromText_PartPatterns(t *testing.T) {
	text := `Part I—Income
Report all sources of income in this section.
Wages from employment should be entered on the first line.
Interest and dividends follow in subsequent lines.
Business income is reported separately on Schedule C.

Part II—Adjustments to Income
Certain deductions are taken above the line to arrive at AGI.
These include educator expenses, IRA contributions, and student loan interest.
Self-employed individuals can also deduct half of self-employment tax.

Part III—Tax and Credits
Calculate your tax using the tax tables or tax computation worksheet.
Apply any credits you are eligible for to reduce your tax liability.
`

	docs := ExtractFromText(text, "Form 1040 Instructions", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) != 3 {
		t.Fatalf("expected 3 chunks from Part patterns, got %d", len(docs))
	}

	// Verify section identifiers
	sections := make(map[string]bool)
	for _, doc := range docs {
		sections[doc.Section] = true
	}
	for _, expected := range []string{"Part I", "Part II", "Part III"} {
		if !sections[expected] {
			t.Errorf("expected section %q in results, got sections: %v", expected, sections)
		}
	}
}

func TestExtractFromText_ShortChunksSkipped(t *testing.T) {
	text := `Line 1. Wages
Enter wages here.
Include all W-2 income from employers and other compensation.

Line 2. Short
Hi.

Line 3. Dividends
Enter your total ordinary dividends from all 1099-DIV forms received.
Include dividends from mutual funds, stocks, and other investments.
`

	docs := ExtractFromText(text, "Form 1040", JurisdictionFederal, DocTypeIRSInstruction)

	// Line 2 content "Hi." is < 50 chars, should be skipped
	for _, doc := range docs {
		if strings.Contains(doc.Section, "Line 2") {
			t.Error("Line 2 chunk should have been skipped (content too short)")
		}
	}

	if len(docs) < 2 {
		t.Fatalf("expected at least 2 chunks (Lines 1 and 3), got %d", len(docs))
	}
}

func TestExtractFromText_TagGeneration(t *testing.T) {
	text := `Line 1. Wages and income
Enter your wages, salaries, and tips from Form W-2.
This includes all compensation for services performed as an employee.
Federal income tax withholding is reported in Box 2 of your W-2.
`

	docs := ExtractFromText(text, "Form 1040", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) == 0 {
		t.Fatal("expected at least 1 document")
	}

	doc := docs[0]
	if len(doc.Tags) == 0 {
		t.Fatal("expected tags to be generated")
	}

	// Should have common tax-related tags
	tagSet := make(map[string]bool)
	for _, tag := range doc.Tags {
		tagSet[tag] = true
	}

	// "income" and "wages" and "withholding" should appear
	expectedTags := []string{"income", "wages", "withholding"}
	for _, expected := range expectedTags {
		if !tagSet[expected] {
			t.Errorf("expected tag %q in tags %v", expected, doc.Tags)
		}
	}

	// Tags should not exceed 8
	if len(doc.Tags) > 8 {
		t.Errorf("expected at most 8 tags, got %d", len(doc.Tags))
	}
}

func TestExtractFromText_SequentialIDs(t *testing.T) {
	text := `Line 1. First section
This is the content of the first section with enough text to pass the minimum.

Line 2. Second section
This is the content of the second section with enough text to pass the minimum.

Line 3. Third section
This is the content of the third section with enough text to pass the minimum.
`

	docs := ExtractFromText(text, "Form 1040", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	// Check sequential IDs
	for i, doc := range docs {
		expectedSuffix := strings.Replace(doc.ID, "", "", 0) // just verify they end with _1, _2, _3
		if !strings.HasSuffix(doc.ID, fmt.Sprintf("_%d", i+1)) {
			t.Errorf("doc[%d] ID = %q, expected sequential suffix _%d (got %s)", i, doc.ID, i+1, expectedSuffix)
		}
	}

	// Verify IDs have proper prefix
	for _, doc := range docs {
		if !strings.HasPrefix(doc.ID, "extract_federal_") {
			t.Errorf("doc ID = %q, expected prefix 'extract_federal_'", doc.ID)
		}
	}
}

func TestExtractFromText_CAJurisdiction(t *testing.T) {
	text := `Part I—Income Adjustments
California adjustments to federal income are reported in this section.
Add back the QBI deduction since California does not conform to IRC 199A.
Social Security benefits should be subtracted from California income.

Part II—Deduction Adjustments
Adjust your federal itemized deductions for California differences.
The SALT deduction cap does not apply for California purposes.
`

	docs := ExtractFromText(text, "Schedule CA (540)", JurisdictionCA, DocTypeFTBInstruction)

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	for _, doc := range docs {
		if doc.Jurisdiction != JurisdictionCA {
			t.Errorf("expected CA jurisdiction, got %s", doc.Jurisdiction)
		}
		if doc.DocType != DocTypeFTBInstruction {
			t.Errorf("expected ftb_instruction doc type, got %s", doc.DocType)
		}
		if !strings.HasPrefix(doc.ID, "extract_ca_") {
			t.Errorf("expected CA ID prefix, got %s", doc.ID)
		}
	}
}

func TestExtractFromText_EmptyInput(t *testing.T) {
	docs := ExtractFromText("", "Test", JurisdictionFederal, DocTypeIRSInstruction)
	if docs != nil {
		t.Errorf("expected nil for empty input, got %d docs", len(docs))
	}

	docs = ExtractFromText("   \n\n  ", "Test", JurisdictionFederal, DocTypeIRSInstruction)
	if docs != nil {
		t.Errorf("expected nil for whitespace-only input, got %d docs", len(docs))
	}
}

func TestExtractFromText_WordLimit(t *testing.T) {
	// Build a chunk with >500 words
	var words []string
	for i := 0; i < 600; i++ {
		words = append(words, "word")
	}
	longContent := strings.Join(words, " ")

	text := "Line 1. Long section\n" + longContent + "\n"

	docs := ExtractFromText(text, "Form 1040", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) == 0 {
		t.Fatal("expected at least 1 document")
	}

	wordCount := len(strings.Fields(docs[0].Content))
	if wordCount > 500 {
		t.Errorf("expected at most 500 words, got %d", wordCount)
	}
}

func TestExtractFromText_GeneralInstructions(t *testing.T) {
	text := `General Instructions
These instructions explain how to fill out the form.
You should read them carefully before beginning the form.
Gather all relevant documents including W-2s and 1099s.

Specific Instructions
The following instructions correspond to specific lines on the form.
Follow them in order as you complete each section of the return.
`

	docs := ExtractFromText(text, "Form 1040", JurisdictionFederal, DocTypeIRSInstruction)

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	foundGeneral := false
	foundSpecific := false
	for _, doc := range docs {
		if strings.Contains(doc.Title, "General Instructions") {
			foundGeneral = true
		}
		if strings.Contains(doc.Title, "Specific Instructions") {
			foundSpecific = true
		}
	}
	if !foundGeneral {
		t.Error("expected to find General Instructions heading")
	}
	if !foundSpecific {
		t.Error("expected to find Specific Instructions heading")
	}
}

