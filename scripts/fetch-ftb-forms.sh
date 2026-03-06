#!/bin/bash
# fetch-ftb-forms.sh — Download blank California FTB PDF forms for a given tax year
set -euo pipefail

YEAR="${1:-2025}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUT_DIR="${PROJECT_ROOT}/internal/pdf/templates/state/ca/${YEAR}"
mkdir -p "$OUT_DIR"

echo "Downloading California FTB forms for tax year $YEAR..."

# Note: FTB URLs change yearly. These are placeholders.
# Uncomment and update when forms are published.
# curl -o "$OUT_DIR/f540.pdf" "https://www.ftb.ca.gov/forms/2025/2025-540.pdf"
# curl -o "$OUT_DIR/f540ca.pdf" "https://www.ftb.ca.gov/forms/2025/2025-540-CA.pdf"

echo "TODO: Update URLs when $YEAR forms are published (typically early January of the following year)"
echo "For now, use text-based export as fallback."
echo ""
echo "Expected output directory: $OUT_DIR"
echo "Expected files:"
echo "  f540.pdf   — Form 540 California Resident Income Tax Return"
echo "  f540ca.pdf — Schedule CA (540) California Adjustments"
echo ""
echo "Once downloaded, pdfcpu will fill forms automatically."
echo "Text-based fallback is used when PDF templates are not found."
