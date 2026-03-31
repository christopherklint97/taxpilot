import { useState, useEffect, useCallback } from "react"
import { useFieldStore } from "@/stores/field-store"
import {
  validateReturn,
  generateMefXml,
  generateCaXml,
  downloadBlob,
} from "@/api/hooks"
import type { ValidateResponse, ValidationResult } from "@/api/types"
import { cn } from "@/lib/utils"
import {
  AlertCircle,
  AlertTriangle,
  Info,
  CheckCircle2,
  Loader2,
  RefreshCw,
  FileDown,
} from "lucide-react"

function SeverityIcon({ severity }: { severity: ValidationResult["severity"] }) {
  switch (severity) {
    case "error":
      return <AlertCircle className="h-4 w-4 shrink-0 text-red-500" />
    case "warning":
      return <AlertTriangle className="h-4 w-4 shrink-0 text-amber-500" />
    case "info":
      return <Info className="h-4 w-4 shrink-0 text-blue-500" />
  }
}

function ResultRow({ result }: { result: ValidationResult }) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-border bg-card px-3 py-2.5">
      <SeverityIcon severity={result.severity} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
              result.severity === "error" && "bg-red-500/10 text-red-600",
              result.severity === "warning" && "bg-amber-500/10 text-amber-600",
              result.severity === "info" && "bg-blue-500/10 text-blue-600",
            )}
          >
            {result.code}
          </span>
          {result.field_key && (
            <span className="truncate text-[11px] text-muted-foreground">
              {result.field_key}
            </span>
          )}
        </div>
        <p className="mt-0.5 text-sm leading-snug text-foreground">
          {result.message}
        </p>
      </div>
    </div>
  )
}

export function ValidationPanel() {
  const returnId = useFieldStore((s) => s.returnId)
  const [result, setResult] = useState<ValidateResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [exporting, setExporting] = useState<string | null>(null)

  const runValidation = useCallback(async () => {
    if (!returnId) return
    setLoading(true)
    setError(null)
    try {
      const res = await validateReturn(returnId)
      setResult(res)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Validation failed")
    } finally {
      setLoading(false)
    }
  }, [returnId])

  useEffect(() => {
    runValidation()
  }, [runValidation])

  const handleExport = async (type: "mef" | "ca") => {
    if (!returnId) return
    setExporting(type)
    try {
      const blob =
        type === "mef"
          ? await generateMefXml(returnId)
          : await generateCaXml(returnId)
      downloadBlob(
        blob,
        `${type === "mef" ? "federal" : "california"}_efile.xml`,
      )
    } catch {
      // Export errors are transient; user can retry
    } finally {
      setExporting(null)
    }
  }

  const errors = result?.results.filter((r) => r.severity === "error") ?? []
  const warnings = result?.results.filter((r) => r.severity === "warning") ?? []
  const infos = result?.results.filter((r) => r.severity === "info") ?? []

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-3xl space-y-6 p-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Validation & Export</h2>
          <button
            onClick={runValidation}
            disabled={loading}
            className="inline-flex items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-xs font-medium text-foreground transition-colors hover:bg-accent disabled:opacity-50"
          >
            {loading ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <RefreshCw className="h-3.5 w-3.5" />
            )}
            Re-validate
          </button>
        </div>

        {/* Summary bar */}
        {loading && !result && (
          <div className="flex items-center justify-center rounded-lg border border-border bg-muted/30 py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <span className="ml-2 text-sm text-muted-foreground">
              Running validation...
            </span>
          </div>
        )}

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900 dark:bg-red-950/30 dark:text-red-400">
            {error}
          </div>
        )}

        {result && (
          <>
            {/* Status badge + counts */}
            <div className="flex items-center gap-4 rounded-lg border border-border bg-muted/30 px-4 py-3">
              <div
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold",
                  result.is_valid
                    ? "bg-emerald-500/10 text-emerald-600"
                    : "bg-red-500/10 text-red-600",
                )}
              >
                {result.is_valid ? (
                  <CheckCircle2 className="h-3.5 w-3.5" />
                ) : (
                  <AlertCircle className="h-3.5 w-3.5" />
                )}
                {result.is_valid ? "Valid" : "Invalid"}
              </div>

              <div className="flex gap-3 text-xs text-muted-foreground">
                {errors.length > 0 && (
                  <span className="flex items-center gap-1 text-red-500">
                    <AlertCircle className="h-3 w-3" />
                    {errors.length} error{errors.length !== 1 && "s"}
                  </span>
                )}
                {warnings.length > 0 && (
                  <span className="flex items-center gap-1 text-amber-500">
                    <AlertTriangle className="h-3 w-3" />
                    {warnings.length} warning
                    {warnings.length !== 1 && "s"}
                  </span>
                )}
                {infos.length > 0 && (
                  <span className="flex items-center gap-1 text-blue-500">
                    <Info className="h-3 w-3" />
                    {infos.length} info
                  </span>
                )}
                {result.results.length === 0 && (
                  <span>No issues found</span>
                )}
              </div>
            </div>

            {/* Grouped results */}
            {errors.length > 0 && (
              <section>
                <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-red-500">
                  Errors
                </h3>
                <div className="space-y-2">
                  {errors.map((r, i) => (
                    <ResultRow key={`err-${i}`} result={r} />
                  ))}
                </div>
              </section>
            )}

            {warnings.length > 0 && (
              <section>
                <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-amber-500">
                  Warnings
                </h3>
                <div className="space-y-2">
                  {warnings.map((r, i) => (
                    <ResultRow key={`warn-${i}`} result={r} />
                  ))}
                </div>
              </section>
            )}

            {infos.length > 0 && (
              <section>
                <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-blue-500">
                  Info
                </h3>
                <div className="space-y-2">
                  {infos.map((r, i) => (
                    <ResultRow key={`info-${i}`} result={r} />
                  ))}
                </div>
              </section>
            )}
          </>
        )}

        {/* Export section */}
        <section className="rounded-lg border border-border bg-card p-4">
          <h3 className="mb-3 text-sm font-semibold">Export</h3>
          <div className="flex flex-wrap gap-3">
            <button
              onClick={() => handleExport("mef")}
              disabled={exporting !== null}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            >
              {exporting === "mef" ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <FileDown className="h-4 w-4" />
              )}
              Federal XML (MeF)
            </button>
            <button
              onClick={() => handleExport("ca")}
              disabled={exporting !== null}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            >
              {exporting === "ca" ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <FileDown className="h-4 w-4" />
              )}
              California XML (FTB)
            </button>
          </div>
        </section>
      </div>
    </div>
  )
}
