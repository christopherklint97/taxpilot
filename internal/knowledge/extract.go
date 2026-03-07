package knowledge

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// sectionPattern matches common IRS/FTB instruction section headings.
var sectionPatterns = []*regexp.Regexp{
	// "Line 1", "Line 1.", "Line 1a.", "Lines 1 through 3"
	regexp.MustCompile(`(?i)^Lines?\s+\d+[a-z]?[\.\s]`),
	// "Part I", "Part I—Title", "Part II -- Something"
	regexp.MustCompile(`(?i)^Part\s+[IVX]+\s*[—\-\.]`),
	regexp.MustCompile(`(?i)^Part\s+[IVX]+\s*$`),
	// "General Instructions", "Specific Instructions", "Special Instructions"
	regexp.MustCompile(`(?i)^(General|Specific|Special)\s+Instructions`),
	// ALL CAPS headings with at least 10 characters
	regexp.MustCompile(`^[A-Z][A-Z\s]{9,}$`),
}

// lineNumberPattern extracts "Line N" for section naming.
var lineNumberPattern = regexp.MustCompile(`(?i)Lines?\s+(\d+[a-z]?)`)

// partPattern extracts "Part I", "Part II", etc.
var partPattern = regexp.MustCompile(`(?i)(Part\s+[IVX]+)`)

// tagKeywords are terms that indicate useful tags when found in content.
var tagKeywords = []string{
	"deduction", "credit", "income", "wages", "tax", "filing",
	"exemption", "withholding", "adjustment", "dependent",
	"capital", "gains", "loss", "business", "self-employment",
	"retirement", "IRA", "401k", "HSA", "medical", "charitable",
	"mortgage", "interest", "dividend", "Social Security",
	"Medicare", "FICA", "AMT", "EITC", "child", "education",
	"estimated", "penalty", "refund", "payment", "conformity",
	"California", "federal", "schedule", "form", "worksheet",
}

// ExtractFromText takes raw text content and chunks it into Documents.
// It splits on section boundaries (Line N, Part I, ALL CAPS headings, etc.),
// generates meaningful titles, auto-generates tags, assigns sequential IDs,
// skips chunks shorter than 50 characters, and limits chunks to ~500 words.
func ExtractFromText(text, source string, jurisdiction Jurisdiction, docType DocumentType) []Document {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	idPrefix := buildIDPrefix(source, jurisdiction)

	type chunk struct {
		title   string
		section string
		lines   []string
	}

	var chunks []chunk
	var current *chunk

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current != nil {
				current.lines = append(current.lines, "")
			}
			continue
		}

		if isHeading(trimmed) {
			// Start a new chunk
			title, section := extractTitleAndSection(trimmed)
			c := chunk{title: title, section: section}
			chunks = append(chunks, c)
			current = &chunks[len(chunks)-1]
		} else {
			if current == nil {
				// Content before any heading -- create an intro chunk
				c := chunk{title: "Introduction", section: "intro"}
				chunks = append(chunks, c)
				current = &chunks[len(chunks)-1]
			}
			current.lines = append(current.lines, line)
		}
	}

	// Convert chunks to Documents
	var docs []Document
	seqNum := 0

	for _, c := range chunks {
		content := normalizeContent(strings.Join(c.lines, "\n"))

		// Skip chunks that are too short
		if len(content) < 50 {
			continue
		}

		// Limit to ~500 words
		content = limitWords(content, 500)

		seqNum++
		doc := Document{
			ID:           fmt.Sprintf("%s_%d", idPrefix, seqNum),
			Title:        fmt.Sprintf("%s - %s", source, c.title),
			Content:      content,
			Source:        source,
			Jurisdiction: jurisdiction,
			DocType:      docType,
			Section:      c.section,
			Tags:         generateTags(c.title, content),
		}
		docs = append(docs, doc)
	}

	return docs
}

// isHeading returns true if a line matches a known section heading pattern.
func isHeading(line string) bool {
	for _, pat := range sectionPatterns {
		if pat.MatchString(line) {
			return true
		}
	}
	return false
}

// extractTitleAndSection derives a title and section identifier from a heading line.
func extractTitleAndSection(line string) (title, section string) {
	title = line
	// Truncate very long titles
	if len(title) > 120 {
		title = title[:120]
	}

	// Try to extract a section identifier
	if m := lineNumberPattern.FindStringSubmatch(line); len(m) > 1 {
		section = "Line " + m[1]
	} else if m := partPattern.FindStringSubmatch(line); len(m) > 1 {
		section = m[1]
	} else {
		// Use the heading itself as the section
		section = strings.TrimSpace(line)
		if len(section) > 60 {
			section = section[:60]
		}
	}

	return title, section
}

// normalizeContent cleans up extracted text: collapses multiple blank lines,
// trims leading/trailing whitespace, and normalizes spaces.
func normalizeContent(text string) string {
	// Collapse runs of whitespace-only lines into single blank lines
	lines := strings.Split(text, "\n")
	var result []string
	prevBlank := true // start as true to trim leading blank lines
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevBlank {
				result = append(result, "")
			}
			prevBlank = true
		} else {
			result = append(result, trimmed)
			prevBlank = false
		}
	}

	// Trim trailing blank lines
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return strings.Join(result, " ")
}

// limitWords truncates text to approximately maxWords words.
func limitWords(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}
	return strings.Join(words[:maxWords], " ")
}

// generateTags extracts relevant tags from the title and content.
func generateTags(title, content string) []string {
	combined := strings.ToLower(title + " " + content)
	seen := make(map[string]bool)
	var tags []string

	for _, kw := range tagKeywords {
		kwLower := strings.ToLower(kw)
		if strings.Contains(combined, kwLower) && !seen[kwLower] {
			seen[kwLower] = true
			tags = append(tags, kwLower)
			if len(tags) >= 8 {
				break
			}
		}
	}

	return tags
}

// buildIDPrefix creates a base ID prefix like "extract_federal_1040".
func buildIDPrefix(source string, jurisdiction Jurisdiction) string {
	// Extract form number/name from source for the ID
	clean := strings.ToLower(source)
	// Remove common prefixes
	for _, prefix := range []string{"irs ", "ftb ", "form ", "schedule "} {
		clean = strings.Replace(clean, prefix, "", 1)
	}
	// Keep only alphanumeric and underscores
	var b strings.Builder
	for _, r := range clean {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune('_')
		}
	}
	idPart := b.String()
	// Trim trailing underscores
	idPart = strings.TrimRight(idPart, "_")
	if idPart == "" {
		idPart = "doc"
	}
	return fmt.Sprintf("extract_%s_%s", jurisdiction, idPart)
}
