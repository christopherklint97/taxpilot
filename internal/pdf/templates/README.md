# PDF Templates

PDF templates are not stored in git (they're large binaries). Download all forms with:

```bash
./scripts/download-templates.sh
```

This downloads 20 blank PDF forms (16 federal from IRS, 4 CA from FTB) used as templates for filling. The export (`[e]` in rollforward or `--export`) uses pdfcpu to fill these AcroForm PDFs with computed values.

If templates are missing, the filler falls back to text export (`.txt` files).
