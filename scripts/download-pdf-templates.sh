#!/usr/bin/env bash
# Download IRS and CA FTB PDF form templates for TaxPilot.
# Usage: ./scripts/download-pdf-templates.sh [year]
# If no year is given, downloads both 2024 and 2025.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_DIR/data/tax_years"

# Map: taxpilot_form_id -> IRS filename (without .pdf)
declare -A FEDERAL_FORMS=(
  [1040]="f1040"
  [schedule_a]="f1040sa"
  [schedule_b]="f1040sb"
  [schedule_c]="f1040sc"
  [schedule_d]="f1040sd"
  [schedule_1]="f1040s1"
  [schedule_2]="f1040s2"
  [schedule_3]="f1040s3"
  [schedule_se]="f1040sse"
  [form_8949]="f8949"
  [form_8889]="f8889"
  [form_8995]="f8995"
  [form_2555]="f2555"
  [form_1116]="f1116"
  [form_8938]="f8938"
  [form_8833]="f8833"
)

# Map: taxpilot_form_id -> FTB form number
declare -A CA_FORMS=(
  [ca_540]="540"
  [ca_schedule_ca]="540-ca"
  [form_3514]="3514"
  [form_3853]="3853"
)

download_federal() {
  local year=$1
  local dest_dir="$DATA_DIR/$year/federal"
  mkdir -p "$dest_dir"

  echo "=== Downloading federal forms for $year ==="
  for form_id in "${!FEDERAL_FORMS[@]}"; do
    local irs_name="${FEDERAL_FORMS[$form_id]}"
    local dest="$dest_dir/${form_id}.pdf"

    if [[ -f "$dest" ]]; then
      echo "  [skip] $form_id — already exists"
      continue
    fi

    # IRS uses /pub/irs-prior/ for past years, /pub/irs-pdf/ for current
    local url="https://www.irs.gov/pub/irs-prior/${irs_name}--${year}.pdf"

    echo -n "  [download] $form_id ($url) ... "
    if curl -fsSL -o "$dest" "$url" 2>/dev/null; then
      echo "OK ($(du -h "$dest" | cut -f1 | xargs))"
    else
      # Try current-year URL as fallback
      url="https://www.irs.gov/pub/irs-pdf/${irs_name}.pdf"
      echo -n "retrying ($url) ... "
      if curl -fsSL -o "$dest" "$url" 2>/dev/null; then
        echo "OK ($(du -h "$dest" | cut -f1 | xargs))"
      else
        rm -f "$dest"
        echo "FAILED"
      fi
    fi
  done
}

download_ca() {
  local year=$1
  local dest_dir="$DATA_DIR/$year/ca"
  mkdir -p "$dest_dir"

  echo "=== Downloading CA FTB forms for $year ==="
  for form_id in "${!CA_FORMS[@]}"; do
    local ftb_name="${CA_FORMS[$form_id]}"
    local dest="$dest_dir/${form_id}.pdf"

    if [[ -f "$dest" ]]; then
      echo "  [skip] $form_id — already exists"
      continue
    fi

    local url="https://www.ftb.ca.gov/forms/${year}/${year}-${ftb_name}.pdf"

    echo -n "  [download] $form_id ($url) ... "
    if curl -fsSL -o "$dest" "$url" 2>/dev/null; then
      echo "OK ($(du -h "$dest" | cut -f1 | xargs))"
    else
      rm -f "$dest"
      echo "FAILED"
    fi
  done
}

# Determine which years to download
if [[ $# -gt 0 ]]; then
  YEARS=("$@")
else
  YEARS=(2024 2025)
fi

for year in "${YEARS[@]}"; do
  download_federal "$year"
  download_ca "$year"
  echo ""
done

echo "Done. Templates saved to $DATA_DIR/"
