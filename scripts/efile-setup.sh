#!/bin/bash
# efile-setup.sh — Interactive guide for e-file provider registration
#
# This script walks through the requirements for becoming an IRS e-file
# provider and CA FTB e-file provider, and saves configuration as JSON
# to ~/.taxpilot/efile.json.
#
# Prerequisites for IRS MeF e-filing:
#   1. Apply for EFIN (Electronic Filing Identification Number) at
#      https://www.irs.gov/e-file-providers/e-file-for-tax-professionals
#   2. Pass Suitability Check (fingerprinting, background check)
#   3. Obtain IRS e-Services account
#   4. Get Strong Authentication certificate (X.509)
#   5. Pass ATS (Application Testing System) certification
#   6. Submit production pilot returns
#
# Prerequisites for CA FTB e-filing:
#   1. Submit Letter of Intent (LOI) to FTB
#   2. Register as e-file provider with FTB
#   3. Pass PATS (Provider Acceptance Testing System) certification
#   4. Obtain FTB provider credentials
#
# Status: This script provides guidance only. Actual registration is a
# multi-week bureaucratic process that must be started well in advance
# of tax filing season.

set -euo pipefail

CONFIG_DIR="$HOME/.taxpilot"
CONFIG_FILE="$CONFIG_DIR/efile.json"

# Ensure config directory exists
mkdir -p "$CONFIG_DIR"

# --- Helper functions ---

# Read a JSON field from the config file
read_json_field() {
    local field="$1"
    if [ -f "$CONFIG_FILE" ]; then
        python3 -c "
import json, sys
try:
    with open('$CONFIG_FILE') as f:
        data = json.load(f)
    print(data.get('$field', ''))
except:
    print('')
" 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# Read a JSON bool field from the config file
read_json_bool() {
    local field="$1"
    if [ -f "$CONFIG_FILE" ]; then
        python3 -c "
import json
try:
    with open('$CONFIG_FILE') as f:
        data = json.load(f)
    print('true' if data.get('$field', False) else 'false')
except:
    print('false')
" 2>/dev/null || echo "false"
    else
        echo "false"
    fi
}

# Write a field to the JSON config file
write_json_field() {
    local field="$1"
    local value="$2"
    local is_bool="${3:-false}"

    if [ -f "$CONFIG_FILE" ]; then
        python3 -c "
import json
with open('$CONFIG_FILE') as f:
    data = json.load(f)
if '$is_bool' == 'true':
    data['$field'] = $value
else:
    data['$field'] = '$value'
with open('$CONFIG_FILE', 'w') as f:
    json.dump(data, f, indent=2)
"
    else
        if [ "$is_bool" = "true" ]; then
            python3 -c "
import json
data = {'$field': $value}
with open('$CONFIG_FILE', 'w') as f:
    json.dump(data, f, indent=2)
"
        else
            python3 -c "
import json
data = {'$field': '$value'}
with open('$CONFIG_FILE', 'w') as f:
    json.dump(data, f, indent=2)
"
        fi
    fi
    chmod 600 "$CONFIG_FILE"
}

# Status indicator
status_icon() {
    if [ "$1" = "true" ] || [ -n "$1" ] && [ "$1" != "false" ] && [ "$1" != "(not set)" ]; then
        echo "[OK]"
    else
        echo "[--]"
    fi
}

echo "==========================================="
echo "  TaxPilot E-File Provider Setup Guide"
echo "==========================================="
echo ""

# --- Show current status if config exists ---
if [ -f "$CONFIG_FILE" ]; then
    echo "--- Current E-File Configuration Status ---"
    echo ""

    cur_efin=$(read_json_field "efin")
    cur_efin_approved=$(read_json_bool "efin_approved")
    cur_cert_path=$(read_json_field "cert_path")
    cur_cert_expiry=$(read_json_field "cert_expiry")
    cur_ats_passed=$(read_json_bool "ats_passed")
    cur_ats_mode=$(read_json_bool "ats_mode")
    cur_ftb_id=$(read_json_field "ftb_provider_id")
    cur_pats_passed=$(read_json_bool "pats_passed")
    cur_pats_mode=$(read_json_bool "pats_mode")

    echo "  Federal (IRS MeF):"
    echo "    $(status_icon "$cur_efin") EFIN:            ${cur_efin:-(not set)}"
    echo "    $(status_icon "$cur_efin_approved") EFIN Approved:   $cur_efin_approved"
    echo "    $(status_icon "$cur_cert_path") Certificate:     ${cur_cert_path:-(not set)}"
    if [ -n "$cur_cert_expiry" ]; then
        echo "    [..] Cert Expiry:     $cur_cert_expiry"
    fi
    echo "    $(status_icon "$cur_ats_passed") ATS Passed:      $cur_ats_passed"
    echo "    [..] ATS Mode:        $cur_ats_mode"

    # Determine federal readiness
    if [ -n "$cur_efin" ] && [ "$cur_efin_approved" = "true" ] && [ -n "$cur_cert_path" ] && [ "$cur_ats_passed" = "true" ]; then
        echo "    >>> Federal Status:   READY"
    else
        echo "    >>> Federal Status:   NOT READY"
    fi

    echo ""
    echo "  California (FTB):"
    echo "    $(status_icon "$cur_ftb_id") FTB Provider ID: ${cur_ftb_id:-(not set)}"
    echo "    $(status_icon "$cur_pats_passed") PATS Passed:     $cur_pats_passed"
    echo "    [..] PATS Mode:       $cur_pats_mode"

    if [ -n "$cur_ftb_id" ] && [ "$cur_pats_passed" = "true" ]; then
        echo "    >>> CA Status:        READY"
    else
        echo "    >>> CA Status:        NOT READY"
    fi

    echo ""
    echo "  Config file: $CONFIG_FILE"
    echo ""
    echo "-------------------------------------------"
    echo ""
fi

echo "This guide walks through the requirements for"
echo "enabling electronic filing with the IRS and CA FTB."
echo ""

echo "--- STEP 1: IRS E-File Provider Registration ---"
echo ""
echo "  1. Visit: https://www.irs.gov/e-file-providers/e-file-for-tax-professionals"
echo "  2. Apply for an EFIN (Electronic Filing Identification Number)"
echo "     Submit Form 8633 with SSN, PTIN, and business information."
echo "  3. Complete the Suitability Check (fingerprinting required)"
echo "     Must be completed within 15 days of EFIN application."
echo "     Results take 2-4 weeks."
echo "  4. Create an IRS e-Services account"
echo "  5. Wait for EFIN approval (typically 4-6 weeks)"
echo ""
read -r -p "Do you have an EFIN? (y/n): " has_efin
if [ "$has_efin" = "y" ]; then
    read -r -p "Enter your EFIN: " efin
    write_json_field "efin" "$efin"
    echo "  EFIN saved to $CONFIG_FILE"

    read -r -p "Has your EFIN been approved by the IRS? (y/n): " efin_approved
    if [ "$efin_approved" = "y" ]; then
        write_json_field "efin_approved" "true" "true"
        echo "  EFIN approval status saved."
    else
        write_json_field "efin_approved" "false" "true"
        echo "  EFIN marked as pending approval."
    fi
else
    echo "  You need an EFIN before you can e-file. Apply at the URL above."
    echo "  The process takes 4-6 weeks. Start early!"
fi

echo ""
echo "--- STEP 2: Strong Authentication Certificate ---"
echo ""
echo "  The IRS MeF system requires a Strong Authentication certificate"
echo "  (X.509) for API access. This is obtained through IRS e-Services"
echo "  after your EFIN is approved."
echo ""
echo "  Certificate must be:"
echo "    - PKCS#12 format (.p12 or .pfx)"
echo "    - Stored securely (not in version control)"
echo "    - Renewed annually"
echo ""
read -r -p "Do you have a certificate file? (y/n): " has_cert
if [ "$has_cert" = "y" ]; then
    read -r -p "Enter path to certificate file (.p12): " cert_path
    if [ -f "$cert_path" ]; then
        write_json_field "cert_path" "$cert_path"
        echo "  Certificate path saved to $CONFIG_FILE"

        read -r -p "Enter certificate expiry date (YYYY-MM-DD): " cert_expiry
        if [ -n "$cert_expiry" ]; then
            write_json_field "cert_expiry" "$cert_expiry"
            echo "  Certificate expiry saved."
        fi
    else
        echo "  WARNING: File not found at $cert_path"
        echo "  Certificate path not saved."
    fi
else
    echo "  You need a certificate before you can e-file."
    echo "  Obtain one through IRS e-Services after EFIN approval."
fi

echo ""
echo "--- STEP 3: ATS Testing ---"
echo ""
echo "  Before production e-filing, you must pass ATS certification:"
echo "    1. Submit test returns to the ATS endpoint"
echo "    2. Receive and process test acknowledgements"
echo "    3. Handle test rejection scenarios"
echo "    4. Demonstrate PIN-based signature handling"
echo "    5. Pass all required test cases"
echo ""
echo "  Run: taxpilot --efile --test-mode"
echo ""
read -r -p "Have you passed ATS certification? (y/n): " ats_passed
if [ "$ats_passed" = "y" ]; then
    write_json_field "ats_passed" "true" "true"
    write_json_field "ats_mode" "false" "true"
    echo "  ATS certification recorded. ATS mode disabled."
else
    write_json_field "ats_passed" "false" "true"
    read -r -p "Enable ATS test mode? (y/n): " enable_ats
    if [ "$enable_ats" = "y" ]; then
        write_json_field "ats_mode" "true" "true"
        echo "  ATS test mode enabled."
    fi
fi

echo ""
echo "--- STEP 4: CA FTB Registration ---"
echo ""
echo "  For California e-filing:"
echo "    1. Submit a Letter of Intent (LOI) to FTB"
echo "       Due by November 1 for the following tax year."
echo "       https://www.ftb.ca.gov/tax-pros/e-file/"
echo "    2. Complete FTB provider registration"
echo "    3. Pass PATS certification testing"
echo "       Must test Form 540, Schedule CA, and supported credits."
echo "    4. Receive FTB provider credentials"
echo ""
read -r -p "Do you have a CA FTB provider ID? (y/n): " has_ftb
if [ "$has_ftb" = "y" ]; then
    read -r -p "Enter your FTB Provider ID: " ftb_id
    write_json_field "ftb_provider_id" "$ftb_id"
    echo "  FTB Provider ID saved to $CONFIG_FILE"

    read -r -p "Have you passed PATS certification? (y/n): " pats_passed
    if [ "$pats_passed" = "y" ]; then
        write_json_field "pats_passed" "true" "true"
        write_json_field "pats_mode" "false" "true"
        echo "  PATS certification recorded. PATS mode disabled."
    else
        write_json_field "pats_passed" "false" "true"
        read -r -p "Enable PATS test mode? (y/n): " enable_pats
        if [ "$enable_pats" = "y" ]; then
            write_json_field "pats_mode" "true" "true"
            echo "  PATS test mode enabled."
        fi
    fi
else
    echo "  You need FTB provider credentials before CA e-filing."
    echo "  Submit LOI by November 1 for the following tax year."
fi

echo ""
echo "--- STEP 5: Configuration Summary ---"
echo ""
echo "  E-file configuration is stored in: $CONFIG_FILE"
echo ""

if [ -f "$CONFIG_FILE" ]; then
    echo "  Current configuration:"
    echo ""
    python3 -c "
import json
with open('$CONFIG_FILE') as f:
    data = json.load(f)
print(json.dumps(data, indent=2))
" 2>/dev/null || cat "$CONFIG_FILE"
    echo ""
fi

echo ""
echo "  For detailed registration documentation, see:"
echo "  docs/efile-provider-guide.md"
echo ""
echo "==========================================="
echo "  Setup guide complete."
echo "==========================================="
