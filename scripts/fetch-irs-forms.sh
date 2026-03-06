#!/bin/bash
# fetch-irs-forms.sh — Download blank IRS PDF forms for a given tax year
set -euo pipefail

YEAR="${1:-2025}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUT_DIR="${PROJECT_ROOT}/internal/pdf/templates/federal/${YEAR}"
mkdir -p "$OUT_DIR"

echo "Downloading IRS forms for tax year $YEAR..."

# Note: IRS URLs change yearly. These are placeholders.
# Uncomment and update when forms are published.
# curl -o "$OUT_DIR/f1040.pdf" "https://www.irs.gov/pub/irs-pdf/f1040.pdf"
# curl -o "$OUT_DIR/f1040s1.pdf" "https://www.irs.gov/pub/irs-pdf/f1040s1.pdf"
# curl -o "$OUT_DIR/f1040s2.pdf" "https://www.irs.gov/pub/irs-pdf/f1040s2.pdf"
# curl -o "$OUT_DIR/f1040s3.pdf" "https://www.irs.gov/pub/irs-pdf/f1040s3.pdf"

echo "TODO: Update URLs when $YEAR forms are published (typically late December of the tax year)"
echo "For now, use text-based export as fallback."
echo ""
echo "Expected output directory: $OUT_DIR"
echo "Expected files:"
echo "  f1040.pdf  — Form 1040 U.S. Individual Income Tax Return"
echo ""
echo "Once downloaded, pdfcpu will fill forms automatically."
echo "Text-based fallback is used when PDF templates are not found."
