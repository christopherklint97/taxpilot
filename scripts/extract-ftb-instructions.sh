#!/usr/bin/env bash
# extract-ftb-instructions.sh — Extract text from CA FTB instruction PDFs and
# generate JSON knowledge base documents for TaxPilot.
#
# Prerequisites:
#   - poppler-utils (for pdftotext): sudo apt install poppler-utils (Linux)
#                                    brew install poppler (macOS)
#
# Usage:
#   ./scripts/extract-ftb-instructions.sh [--year 2025] [--skip-download]

set -euo pipefail

# --- Defaults ---
YEAR="2025"
SKIP_DOWNLOAD=false

# FTB forms typically lag by a year; use 2024 forms as 2025 may not be available
FTB_FORM_YEAR="2024"

# --- Parse flags ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --year)
      YEAR="$2"
      shift 2
      ;;
    --federal-only)
      echo "Skipping CA extraction (--federal-only specified)."
      echo "Run scripts/extract-instructions.sh for federal content."
      exit 0
      ;;
    --ca-only)
      # No-op, we are CA-only by default
      shift
      ;;
    --skip-download)
      SKIP_DOWNLOAD=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--year YYYY] [--skip-download]"
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
PDF_DIR="${PROJECT_DIR}/data/pdfs/ca"
OUTPUT_DIR="${PROJECT_DIR}/data/knowledge/ca"

mkdir -p "${PDF_DIR}"
mkdir -p "${OUTPUT_DIR}"

# --- CA FTB PDF URLs ---
declare -A CA_PDFS=(
  ["540_booklet"]="https://www.ftb.ca.gov/forms/${FTB_FORM_YEAR}/${FTB_FORM_YEAR}-540-booklet.pdf"
  ["540_ca_instructions"]="https://www.ftb.ca.gov/forms/${FTB_FORM_YEAR}/${FTB_FORM_YEAR}-540-ca-instructions.pdf"
)

# --- Form name mappings ---
declare -A FORM_NAMES=(
  ["540_booklet"]="Form 540"
  ["540_ca_instructions"]="Schedule CA (540)"
)

echo "=== CA FTB Instruction PDF Extraction ==="
echo "Tax year: ${YEAR} (using ${FTB_FORM_YEAR} FTB forms)"
echo "PDF directory: ${PDF_DIR}"
echo "Output directory: ${OUTPUT_DIR}"
echo ""

# --- Download PDFs ---
if [[ "${SKIP_DOWNLOAD}" == "false" ]]; then
  echo "--- Downloading FTB instruction PDFs ---"
  for key in "${!CA_PDFS[@]}"; do
    url="${CA_PDFS[$key]}"
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

JSON_DOCS="["
FIRST_DOC=true
DOC_COUNT=0

for key in "${!CA_PDFS[@]}"; do
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
  chunk_id=0
  current_title=""
  current_content=""
  current_section=""

  while IFS= read -r line; do
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
    # Match section-style headings like "Filing Status", "Income", "Deductions"
    elif echo "${line}" | grep -qiE '^\s*(Filing Status|Exemptions|Income|Deductions|Credits|Tax|Payments|Refund|Amount You Owe)'; then
      is_heading=true
      new_title=$(echo "${line}" | sed 's/^[[:space:]]*//' | head -c 120)
      new_section="${new_title}"
    # Match ALL CAPS headings (at least 3 words)
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

        json_content=$(echo "${truncated}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")
        json_title=$(echo "${current_title}" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read().strip()))")

        tags_lower=$(echo "${current_title} ${form_name} california" | tr '[:upper:]' '[:lower:]' | tr -cs '[:alnum:]' ' ')
        tag_array=$(python3 -c "
import json
words = '${tags_lower}'.split()
tags = list(dict.fromkeys(w for w in words if len(w) > 2))[:8]
print(json.dumps(tags))
")

        chunk_id=$((chunk_id + 1))
        DOC_COUNT=$((DOC_COUNT + 1))
        doc_id="extract_ca_${key}_${chunk_id}"

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
    \"source\": \"FTB ${form_name} Instructions\",
    \"jurisdiction\": \"ca\",
    \"doc_type\": \"ftb_instruction\",
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

    tags_lower=$(echo "${current_title} ${form_name} california" | tr '[:upper:]' '[:lower:]' | tr -cs '[:alnum:]' ' ')
    tag_array=$(python3 -c "
import json
words = '${tags_lower}'.split()
tags = list(dict.fromkeys(w for w in words if len(w) > 2))[:8]
print(json.dumps(tags))
")

    chunk_id=$((chunk_id + 1))
    DOC_COUNT=$((DOC_COUNT + 1))
    doc_id="extract_ca_${key}_${chunk_id}"

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
    \"source\": \"FTB ${form_name} Instructions\",
    \"jurisdiction\": \"ca\",
    \"doc_type\": \"ftb_instruction\",
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
echo "Total CA documents extracted: ${DOC_COUNT}"
echo "Output: ${OUTPUT_FILE}"
