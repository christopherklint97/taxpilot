#!/bin/bash
# Downloads blank IRS and CA FTB PDF form templates for filling.
# Usage: ./scripts/download-templates.sh [YEAR]
# Default year: 2025

set -e

YEAR="${1:-2025}"
FEDERAL_DIR="internal/pdf/templates/federal/${YEAR}"
CA_DIR="internal/pdf/templates/state/ca/${YEAR}"

mkdir -p "$FEDERAL_DIR" "$CA_DIR"

echo "Downloading ${YEAR} federal forms from IRS..."
curl -sL -o "$FEDERAL_DIR/f1040.pdf"       "https://www.irs.gov/pub/irs-pdf/f1040.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_1.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040s1.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_2.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040s2.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_3.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040s3.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_a.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040sa.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_b.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040sb.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_c.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040sc.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_d.pdf"   "https://www.irs.gov/pub/irs-pdf/f1040sd.pdf"
curl -sL -o "$FEDERAL_DIR/schedule_se.pdf"  "https://www.irs.gov/pub/irs-pdf/f1040sse.pdf"
curl -sL -o "$FEDERAL_DIR/form_8949.pdf"    "https://www.irs.gov/pub/irs-pdf/f8949.pdf"
curl -sL -o "$FEDERAL_DIR/f8889.pdf"        "https://www.irs.gov/pub/irs-pdf/f8889.pdf"
curl -sL -o "$FEDERAL_DIR/f8995.pdf"        "https://www.irs.gov/pub/irs-pdf/f8995.pdf"
curl -sL -o "$FEDERAL_DIR/f2555.pdf"        "https://www.irs.gov/pub/irs-pdf/f2555.pdf"
curl -sL -o "$FEDERAL_DIR/form_1116.pdf"    "https://www.irs.gov/pub/irs-pdf/f1116.pdf"
curl -sL -o "$FEDERAL_DIR/form_8938.pdf"    "https://www.irs.gov/pub/irs-pdf/f8938.pdf"
curl -sL -o "$FEDERAL_DIR/form_8833.pdf"    "https://www.irs.gov/pub/irs-pdf/f8833.pdf"

# CA FTB uses year-specific URLs
# Note: FTB publishes forms at /forms/YYYY/YYYY-formname.pdf
# Current year forms may also be at the non-year URL
CA_YEAR="${YEAR}"
echo "Downloading ${YEAR} CA forms from FTB..."
curl -sL -o "$CA_DIR/f540.pdf"              "https://www.ftb.ca.gov/forms/${CA_YEAR}/${CA_YEAR}-540.pdf"
curl -sL -o "$CA_DIR/schedule_ca.pdf"       "https://www.ftb.ca.gov/forms/${CA_YEAR}/${CA_YEAR}-540-ca.pdf"
curl -sL -o "$CA_DIR/form_3514.pdf"         "https://www.ftb.ca.gov/forms/${CA_YEAR}/${CA_YEAR}-3514.pdf"
curl -sL -o "$CA_DIR/form_3853.pdf"         "https://www.ftb.ca.gov/forms/${CA_YEAR}/${CA_YEAR}-3853.pdf"

TOTAL=$(find "$FEDERAL_DIR" "$CA_DIR" -name '*.pdf' 2>/dev/null | wc -l | tr -d ' ')
echo "Done. ${TOTAL} PDF templates downloaded for tax year ${YEAR}."
