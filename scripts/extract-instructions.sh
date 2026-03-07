#!/usr/bin/env bash
# extract-instructions.sh — Extract text from IRS instruction PDFs and generate
# JSON knowledge base documents for TaxPilot.
#
# Prerequisites:
#   - poppler-utils (for pdftotext): sudo apt install poppler-utils (Linux)
#                                    brew install poppler (macOS)
#
# Usage:
#   ./scripts/extract-instructions.sh [--year 2025] [--federal-only] [--ca-only] [--skip-download]

set -euo pipefail

# --- Defaults ---
YEAR="2025"
FEDERAL_ONLY=false
CA_ONLY=false
SKIP_DOWNLOAD=false

# --- Parse flags ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --year)
      YEAR="$2"
      shift 2
      ;;
    --federal-only)
      FEDERAL_ONLY=true
      shift
      ;;
    --ca-only)
      CA_ONLY=true
      shift
      ;;
    --skip-download)
      SKIP_DOWNLOAD=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--year YYYY] [--federal-only] [--ca-only] [--skip-download]"
      exit 1
      ;;
  esac
done

# --- Check for pdftotext ---
if ! command -v pdftotext &>/dev/null; then
  echo "ERROR: pdftotext is not installed."
  echo ""
  echo "Install poppler-utils to get pdftotext:"
  echo "  Linux (Debian/Ubuntu): sudo apt install poppler-utils"
  echo "  Linux (Fedora/RHEL):   sudo dnf install poppler-utils"
  echo "  macOS (Homebrew):      brew install poppler"
  echo "  macOS (MacPorts):      sudo port install poppler"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PDF_DIR="${PROJECT_DIR}/data/pdfs/federal"
OUTPUT_DIR="${PROJECT_DIR}/data/knowledge/federal"

mkdir -p "${PDF_DIR}"
mkdir -p "${OUTPUT_DIR}"
mkdir -p "${PROJECT_DIR}/data/knowledge/ca"

# --- Federal PDF URLs ---
declare -A FEDERAL_PDFS=(
  ["i1040"]="https://www.irs.gov/pub/irs-pdf/i1040gi.pdf"
  ["i1040sca"]="https://www.irs.gov/pub/irs-pdf/i1040sca.pdf"
  ["i1040sc"]="https://www.irs.gov/pub/irs-pdf/i1040sc.pdf"
  ["i1040sd"]="https://www.irs.gov/pub/irs-pdf/i1040sd.pdf"
  ["i1040sse"]="https://www.irs.gov/pub/irs-pdf/i1040sse.pdf"
  ["i8889"]="https://www.irs.gov/pub/irs-pdf/i8889.pdf"
  ["i8949"]="https://www.irs.gov/pub/irs-pdf/i8949.pdf"
)

# --- Form name mappings for human-readable titles ---
declare -A FORM_NAMES=(
  ["i1040"]="Form 1040"
  ["i1040sca"]="Schedule A"
  ["i1040sc"]="Schedule C"
  ["i1040sd"]="Schedule D"
  ["i1040sse"]="Schedule SE"
  ["i8889"]="Form 8889"
  ["i8949"]="Form 8949"
)

if [[ "${CA_ONLY}" == "true" ]]; then
  echo "Skipping federal extraction (--ca-only specified)."
  echo "Run scripts/extract-ftb-instructions.sh for CA content."
  exit 0
fi

echo "=== IRS Instruction PDF Extraction ==="
echo "Tax year: ${YEAR}"
echo "PDF directory: ${PDF_DIR}"
echo "Output directory: ${OUTPUT_DIR}"
echo ""

# --- Download PDFs ---
if [[ "${SKIP_DOWNLOAD}" == "false" ]]; then
  echo "--- Downloading IRS instruction PDFs ---"
  for key in "${!FEDERAL_PDFS[@]}"; do
    url="${FEDERAL_PDFS[$key]}"
    dest="${PDF_DIR}/${key}.pdf"
    if [[ -f "${dest}" ]]; then
      echo "  [skip] ${key}.pdf already exists"
    else
      echo "  [download] ${key}.pdf from ${url}"
      if curl -sSfL -o "${dest}" "${url}"; then
        echo "    OK"
      else
        echo "    WARNING: Failed to download ${key}.pdf (URL may not be available yet)"
        rm -f "${dest}"
      fi
    fi
  done
  echo ""
else
  echo "--- Skipping download (--skip-download) ---"
  echo ""
fi

# --- Extract and chunk ---
echo "--- Extracting text and chunking ---"

# JSON output array
JSON_DOCS="["
FIRST_DOC=true
DOC_COUNT=0

for key in "${!FEDERAL_PDFS[@]}"; do
  pdf="${PDF_DIR}/${key}.pdf"
  form_name="${FORM_NAMES[$key]}"

  if [[ ! -f "${pdf}" ]]; then
    echo "  [skip] ${key}.pdf not found"
    continue
  fi

  echo "  [extract] ${key}.pdf (${form_name})"

  # Extract text using pdftotext with layout preservation
  raw_text=$(pdftotext -layout "${pdf}" - 2>/dev/null || true)

  if [[ -z "${raw_text}" ]]; then
    echo "    WARNING: No text extracted from ${key}.pdf"
    continue
  fi

  # Split text into chunks by section headings
  # We look for patterns like "Line 1", "Part I", "General Instructions", etc.
  chunk_id=0
  current_title=""
  current_content=""
  current_section=""

  while IFS= read -r line; do
    # Detect section headings
    is_heading=false
    new_title=""
    new_section=""

    # Match "Line N" or "Lines N through M" patterns
    if echo "${line}" | grep -qiE '^\s*Lines?\s+[0-9]+'; then
      is_heading=true
      new_title=$(echo "${line}" | sed 's/^[[:space:]]*//' | head -c 120)
      new_section=$(echo "${line}" | grep -oiE 'Lines?\s+[0-9]+[a-z]?' | head -1)
    # Match "Part I", "Part II", etc.
    elif echo "${line}" | grep -qiE '^\s*Part\s+[IVX]+'; then
      is_heading=true
      new_title=$(echo "${line}" | sed 's/^[[:space:]]*//' | head -c 120)
      new_section=$(echo "${line}" | grep -oiE 'Part\s+[IVX]+' | head -1)
    # Match "General Instructions", "Specific Instructions", etc.
    elif echo "${line}" | grep -qiE '^\s*(General|Specific|Special)\s+Instructions'; then
      is_heading=true
      new_title=$(echo "${line}" | sed 's/^[[:space:]]*//' | head -c 120)
      new_section=$(echo "${line}" | grep -oiE '(General|Specific|Special)\s+Instructions' | head -1)
    # Match ALL CAPS headings (at least 3 words, all caps)
    elif echo "${line}" | grep -qE '^\s*[A-Z][A-Z ]{10,}$'; then
      is_heading=true
      new_title=$(echo "${line}" | sed 's/^[[:space:]]*//' | head -c 120)
      new_section="${new_title}"
    fi

    if [[ "${is_heading}" == "true" && -n "${new_title}" ]]; then
      # Save previous chunk if it has enough content
      if [[ -n "${current_title}" && ${#current_content} -ge 50 ]]; then
        # Truncate to ~500 words
        truncated=$(echo "${current_content}" | tr '\n' ' ' | sed 's/  */ /g' | awk '{
          n = split($0, words, " ")
          out = ""
          for (i = 1; i <= n && i <= 500; i++) {
            if (i > 1) out = out " "
            out = out words[i]
          }
          print out
        }')

        # Escape for JSON
        json_content=$(echo "${truncated}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")
        json_title=$(echo "${current_title}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")

        # Generate tags from title and form name
        tags_lower=$(echo "${current_title} ${form_name}" | tr '[:upper:]' '[:lower:]' | tr -cs '[:alnum:]' ' ')
        tag_array=$(python3 -c "
import json
words = '${tags_lower}'.split()
# Filter short words and deduplicate
tags = list(dict.fromkeys(w for w in words if len(w) > 2))[:8]
print(json.dumps(tags))
")

        chunk_id=$((chunk_id + 1))
        DOC_COUNT=$((DOC_COUNT + 1))
        doc_id="extract_federal_${key}_${chunk_id}"

        if [[ "${FIRST_DOC}" == "true" ]]; then
          FIRST_DOC=false
        else
          JSON_DOCS="${JSON_DOCS},"
        fi

        JSON_DOCS="${JSON_DOCS}
  {
    \"id\": \"${doc_id}\",
    \"title\": ${json_title},
    \"content\": ${json_content},
    \"source\": \"IRS ${form_name} Instructions\",
    \"jurisdiction\": \"federal\",
    \"doc_type\": \"irs_instruction\",
    \"section\": \"${current_section}\",
    \"tags\": ${tag_array},
    \"tax_year\": ${YEAR}
  }"
      fi

      current_title="${new_title}"
      current_content=""
      current_section="${new_section}"
    else
      current_content="${current_content}
${line}"
    fi
  done <<< "${raw_text}"

  # Save final chunk
  if [[ -n "${current_title}" && ${#current_content} -ge 50 ]]; then
    truncated=$(echo "${current_content}" | tr '\n' ' ' | sed 's/  */ /g' | awk '{
      n = split($0, words, " ")
      out = ""
      for (i = 1; i <= n && i <= 500; i++) {
        if (i > 1) out = out " "
        out = out words[i]
      }
      print out
    }')

    json_content=$(echo "${truncated}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")
    json_title=$(echo "${current_title}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")

    tags_lower=$(echo "${current_title} ${form_name}" | tr '[:upper:]' '[:lower:]' | tr -cs '[:alnum:]' ' ')
    tag_array=$(python3 -c "
import json
words = '${tags_lower}'.split()
tags = list(dict.fromkeys(w for w in words if len(w) > 2))[:8]
print(json.dumps(tags))
")

    chunk_id=$((chunk_id + 1))
    DOC_COUNT=$((DOC_COUNT + 1))
    doc_id="extract_federal_${key}_${chunk_id}"

    if [[ "${FIRST_DOC}" == "true" ]]; then
      FIRST_DOC=false
    else
      JSON_DOCS="${JSON_DOCS},"
    fi

    JSON_DOCS="${JSON_DOCS}
  {
    \"id\": \"${doc_id}\",
    \"title\": ${json_title},
    \"content\": ${json_content},
    \"source\": \"IRS ${form_name} Instructions\",
    \"jurisdiction\": \"federal\",
    \"doc_type\": \"irs_instruction\",
    \"section\": \"${current_section}\",
    \"tags\": ${tag_array},
    \"tax_year\": ${YEAR}
  }"
  fi

  echo "    Extracted ${chunk_id} chunks from ${form_name}"
done

JSON_DOCS="${JSON_DOCS}
]"

# --- Write output ---
OUTPUT_FILE="${OUTPUT_DIR}/instructions.json"
echo "${JSON_DOCS}" > "${OUTPUT_FILE}"

echo ""
echo "=== Done ==="
echo "Total documents extracted: ${DOC_COUNT}"
echo "Output: ${OUTPUT_FILE}"

if [[ "${FEDERAL_ONLY}" != "true" ]]; then
  echo ""
  echo "Run scripts/extract-ftb-instructions.sh for CA FTB instructions."
fi
