#!/usr/bin/env bash
# extract-instructions.sh — Extract text from IRS/FTB instruction PDFs
#
# This script is a stub for the content extraction pipeline. When real
# IRS and FTB PDF instruction booklets are available, this script will:
#
# 1. Download instruction PDFs from IRS.gov and FTB.ca.gov (or read from data/pdfs/)
# 2. Extract text from each PDF using pdftotext or a Go-based extractor
# 3. Split the text into meaningful chunks (by section/topic)
# 4. Generate JSON document files suitable for the knowledge store
# 5. Write output to data/knowledge/{federal,ca}/*.json
#
# Prerequisites:
#   - poppler-utils (for pdftotext): brew install poppler
#   - Or use the Go-based PDF text extractor in internal/pdf/
#
# Usage:
#   ./scripts/extract-instructions.sh [--year 2025] [--federal-only] [--ca-only]
#
# For now, the knowledge base is seeded with hand-curated content in
# internal/knowledge/seed.go. This script will supplement that with
# full-text extraction from official publications when available.

set -euo pipefail

YEAR="${1:-2025}"
DATA_DIR="data/knowledge"

echo "extract-instructions.sh — Content extraction pipeline (stub)"
echo ""
echo "This script is not yet implemented. The knowledge base currently"
echo "uses seed documents defined in internal/knowledge/seed.go."
echo ""
echo "When IRS/FTB instruction PDFs are available in data/pdfs/, this"
echo "script will extract and chunk them into ${DATA_DIR}/."
echo ""
echo "Target tax year: ${YEAR}"
echo "Output directory: ${DATA_DIR}/"
