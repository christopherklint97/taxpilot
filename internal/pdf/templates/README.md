# PDF Templates

PDF templates are not stored in git (they're large binaries). Download them with:

```bash
# Federal Form 1040
mkdir -p internal/pdf/templates/federal/2025
curl -sL -o internal/pdf/templates/federal/2025/f1040.pdf \
  "https://www.irs.gov/pub/irs-pdf/f1040.pdf"

# California Form 540
mkdir -p internal/pdf/templates/state/ca/2025
curl -sL -o internal/pdf/templates/state/ca/2025/f540.pdf \
  "https://www.ftb.ca.gov/forms/2025/2025-540.pdf"
```

The filler automatically falls back to text export when PDF templates are missing.
