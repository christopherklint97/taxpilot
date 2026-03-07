# TaxPilot E-File Provider Registration Guide

This document covers the IRS and CA FTB e-file provider registration process
required before TaxPilot can submit production e-file returns. This is a
multi-month bureaucratic process that must be completed well in advance of
tax filing season.

---

## IRS MeF Provider Registration

### Step 1: IRS e-Services Account

Create an account at https://www.irs.gov/e-file-providers/e-file-for-tax-professionals

Requirements:
- Valid SSN
- PTIN (Preparer Tax Identification Number) -- apply at https://www.irs.gov/tax-professionals/ptin-requirements-for-tax-return-preparers
- Personal identity verification

### Step 2: EFIN Application (Form 8633)

Apply for an EFIN (Electronic Filing Identification Number) via IRS Form 8633,
"Application to Participate in the IRS e-file Program."

Required information:
- SSN and PTIN of the responsible official
- Business name, address, and EIN
- Type of e-file provider (ERO, software developer, etc.)
- Description of software and supported forms

Processing time: 4-6 weeks after submission.

### Step 3: Suitability Check

A suitability check including fingerprinting and an FBI background check is
required for all principals and responsible officials.

- Fingerprinting must be completed within 15 calendar days of submitting
  the EFIN application.
- Use an IRS-approved fingerprint vendor or local law enforcement.
- Results typically take 2-4 weeks to process.
- A criminal record may result in denial.

### Step 4: Strong Authentication Certificate

After EFIN approval, obtain an X.509 digital certificate through IRS e-Services.

Certificate details:
- Format: PKCS#12 (.p12 or .pfx)
- Used to authenticate API calls to the MeF system
- Must be renewed annually before expiration
- Store securely -- NEVER commit to version control
- Configure the path in `~/.taxpilot/efile.json` (see Configuration below)

### Step 5: ATS Certification (Application Testing System)

Before production filing, you must pass ATS certification by submitting test
returns to the IRS ATS endpoint.

Requirements:
- Submit test returns covering all form types TaxPilot supports (1040, schedules)
- Successfully receive acknowledgements for accepted returns
- Handle rejection scenarios (invalid SSN, math errors, duplicate filing)
- Handle error cases (malformed XML, schema violations)
- Demonstrate PIN-based signature handling (Form 8879 flow)

To run TaxPilot in ATS mode:
```bash
taxpilot --efile --test-mode
```

Or set `ats_mode: true` in `~/.taxpilot/efile.json`.

### Step 6: Production Pilot

After ATS certification, the IRS monitors the first 5-10 production returns
before granting full production access.

- Returns are reviewed for accuracy and compliance
- Any issues may require re-certification
- Pilot period typically lasts 1-2 weeks

---

## CA FTB Provider Registration

### Step 1: Letter of Intent (LOI)

Submit a Letter of Intent to the FTB e-file unit.

- Due by November 1 for the following tax year's filing season
- Submit to: CA FTB e-Programs Unit, https://www.ftb.ca.gov/tax-pros/e-file/
- Include: company information, software description, list of supported CA forms

### Step 2: FTB Provider Registration

Complete the FTB provider registration form after LOI acceptance.

- Receive a provider ID for API authentication
- Provide technical contact information
- Agree to FTB e-file provider terms and conditions

### Step 3: PATS Certification (Provider Acceptance Testing System)

Similar to IRS ATS but for California forms.

Required tests:
- Form 540 (California Resident Income Tax Return)
- Schedule CA (California Adjustments)
- Supported credits and schedules
- Conformity edge cases (items where CA differs from federal treatment)
- Accepted, rejected, and error acknowledgement handling

To run TaxPilot in PATS mode:
```bash
taxpilot --efile --state-only --test-mode
```

Or set `pats_mode: true` in `~/.taxpilot/efile.json`.

### Step 4: Provider Credentials

After PATS certification, receive:
- FTB Provider ID
- API credentials for production submission
- Access to the FTB production e-file endpoint

---

## Timeline

Plan to start the registration process at least 6 months before tax filing
season opens (typically January 27).

| Phase                          | Duration     | Deadline           |
|--------------------------------|--------------|--------------------|
| IRS e-Services account         | 1-2 weeks    | July               |
| EFIN application + suitability | 4-6 weeks    | August              |
| Certificate issuance           | 1-2 weeks    | September           |
| ATS certification              | 2-4 weeks    | October             |
| FTB LOI submission             | --           | November 1          |
| FTB registration               | 2-4 weeks    | November            |
| PATS certification             | 2-4 weeks    | December            |
| Production pilot (IRS)         | 1-2 weeks    | Late January        |

Total elapsed time: 5-7 months minimum.

---

## Configuration

TaxPilot reads e-file provider configuration from `~/.taxpilot/efile.json`:

```json
{
  "efin": "123456",
  "cert_path": "/path/to/cert.p12",
  "ats_mode": false,
  "ftb_provider_id": "XXXX",
  "pats_mode": false,
  "efin_approved": true,
  "ats_passed": true,
  "pats_passed": true,
  "cert_expiry": "2027-03-01"
}
```

Field descriptions:

| Field             | Description                                         |
|-------------------|-----------------------------------------------------|
| `efin`            | IRS Electronic Filing Identification Number         |
| `cert_path`       | Path to PKCS#12 certificate file (.p12)             |
| `ats_mode`        | If true, submit to ATS test endpoint instead of production |
| `ftb_provider_id` | CA FTB provider identification number               |
| `pats_mode`       | If true, submit to PATS test endpoint instead of production |
| `efin_approved`   | Set to true after EFIN is approved by the IRS       |
| `ats_passed`      | Set to true after passing ATS certification         |
| `pats_passed`     | Set to true after passing PATS certification        |
| `cert_expiry`     | Certificate expiration date (YYYY-MM-DD)            |

Notes:
- The certificate password is prompted at runtime and never stored on disk.
- Never commit `efile.json` or the certificate file to version control.
- Run `scripts/efile-setup.sh` for an interactive setup wizard.

### Checking Status

Use the setup script to check your current registration status:
```bash
./scripts/efile-setup.sh
```

Or use TaxPilot directly:
```bash
taxpilot --efile-status
```

---

## Security Considerations

1. The PKCS#12 certificate file should have restrictive permissions (`chmod 600`).
2. The certificate password is never written to disk -- it is prompted at runtime.
3. The `efile.json` file should have restrictive permissions (`chmod 600`).
4. Add `efile.json` and `*.p12` to `.gitignore` (already configured).
5. For CI/CD, use environment variables or a secrets manager for credentials.
