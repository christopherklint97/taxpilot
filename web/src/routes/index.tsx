import { useEffect, useRef, useState } from "react"
import { useNavigate } from "@tanstack/react-router"
import { useReturnStore } from "@/stores/return-store"
import { rollforward } from "@/api/hooks"
import {
  FileText,
  Plus,
  Trash2,
  ArrowRight,
  Upload,
  Copy,
} from "lucide-react"

export function DashboardPage() {
  const { returns, loading, fetchReturns, createReturn, deleteReturn } =
    useReturnStore()
  const navigate = useNavigate()
  const [creating, setCreating] = useState(false)
  const [importing, setImporting] = useState(false)
  const [rollingForward, setRollingForward] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    fetchReturns()
  }, [fetchReturns])

  const handleCreate = async () => {
    setCreating(true)
    setError(null)
    try {
      const ret = await createReturn(2025)
      navigate({ to: "/returns/$returnId", params: { returnId: ret.id } })
    } catch (e) {
      setError((e as Error).message || "Failed to create return")
    } finally {
      setCreating(false)
    }
  }

  const handleImport = async (files: FileList | null) => {
    if (!files || files.length === 0) return
    setImporting(true)
    setError(null)
    try {
      const ret = await createReturn(2024)
      for (const file of Array.from(files)) {
        const formData = new FormData()
        formData.append("file", file)
        const res = await fetch(`/api/returns/${ret.id}/pdf/upload`, {
          method: "POST",
          body: formData,
        })
        if (!res.ok) {
          const text = await res.text()
          throw new Error(`Upload failed: ${res.status} ${text}`)
        }
      }
      await fetchReturns()
      navigate({ to: "/returns/$returnId", params: { returnId: ret.id } })
    } catch (e) {
      setError((e as Error).message || "Failed to import PDF")
    } finally {
      setImporting(false)
      if (fileInputRef.current) fileInputRef.current.value = ""
    }
  }

  const handleRollforward = async (sourceId: string) => {
    setRollingForward(sourceId)
    setError(null)
    try {
      const result = await rollforward(sourceId)
      await fetchReturns()
      navigate({
        to: "/returns/$returnId",
        params: { returnId: result.return_id },
      })
    } catch (e) {
      setError((e as Error).message || "Failed to roll forward")
    } finally {
      setRollingForward(null)
    }
  }

  // Sort: newest year first
  const sorted = [...returns].sort((a, b) => b.tax_year - a.tax_year)

  // Check if we already have a 2025 return
  const has2025 = returns.some((r) => r.tax_year === 2025)
  const source2024 = returns.find((r) => r.tax_year === 2024)

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Tax Returns</h2>
          <p className="text-sm text-muted-foreground">
            Import your 2024 return, then roll forward to 2025
          </p>
        </div>
        <div className="flex items-center gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf"
            multiple
            className="hidden"
            onChange={(e) => handleImport(e.target.files)}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={importing}
            className="inline-flex items-center gap-2 rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-accent disabled:opacity-50"
          >
            <Upload className="h-4 w-4" />
            {importing ? "Importing..." : "Import 2024 PDF"}
          </button>
          <button
            onClick={handleCreate}
            disabled={creating}
            className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Plus className="h-4 w-4" />
            {creating ? "Creating..." : "New 2025 Return"}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Quick start: rollforward CTA if 2024 exists but no 2025 */}
      {source2024 && !has2025 && (
        <div className="mb-6 rounded-lg border-2 border-primary/30 bg-primary/5 p-5">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-semibold">
                Ready to file for 2025?
              </h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Copy your 2024 values into a new 2025 return. You can
                review and edit everything before filing.
              </p>
            </div>
            <button
              onClick={() => handleRollforward(source2024.id)}
              disabled={rollingForward === source2024.id}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              <Copy className="h-4 w-4" />
              {rollingForward === source2024.id
                ? "Copying..."
                : "Copy 2024 \u2192 2025"}
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <div className="py-12 text-center text-muted-foreground">
          Loading...
        </div>
      ) : returns.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border p-12 text-center">
          <FileText className="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
          <h3 className="mb-2 text-lg font-medium">No tax returns yet</h3>
          <p className="mb-6 text-sm text-muted-foreground">
            Start by importing your 2024 PDF, then roll forward to 2025.
          </p>
          <div className="flex items-center justify-center gap-3">
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={importing}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              <Upload className="h-4 w-4" />
              {importing ? "Importing..." : "Import 2024 PDF"}
            </button>
            <span className="text-xs text-muted-foreground">or</span>
            <button
              onClick={handleCreate}
              disabled={creating}
              className="inline-flex items-center gap-2 rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-accent"
            >
              <Plus className="h-4 w-4" />
              {creating ? "Creating..." : "Start fresh (2025)"}
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {sorted.map((ret) => (
            <div
              key={ret.id}
              className="flex items-center justify-between rounded-lg border border-border p-4 transition-colors hover:bg-accent/50 cursor-pointer"
              onClick={() =>
                navigate({
                  to: "/returns/$returnId",
                  params: { returnId: ret.id },
                })
              }
            >
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-md bg-muted text-sm font-bold">
                  {String(ret.tax_year).slice(-2)}
                </div>
                <div>
                  <p className="font-medium">
                    {ret.tax_year} Tax Return
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {ret.filing_status
                      ? formatFilingStatus(ret.filing_status)
                      : "Not started"}{" "}
                    &middot; {ret.state_code} &middot; Updated{" "}
                    {new Date(ret.updated_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                {/* Only show rollforward for non-2025 returns */}
                {ret.tax_year < 2025 && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleRollforward(ret.id)
                    }}
                    disabled={rollingForward === ret.id}
                    className="inline-flex items-center gap-1 rounded-md px-2 py-1.5 text-xs text-muted-foreground hover:bg-primary/10 hover:text-primary disabled:opacity-50"
                    title={`Copy to ${ret.tax_year + 1}`}
                  >
                    <ArrowRight className="h-3.5 w-3.5" />
                    <span className="hidden sm:inline">
                      {rollingForward === ret.id
                        ? "Copying..."
                        : `Copy to ${ret.tax_year + 1}`}
                    </span>
                  </button>
                )}
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    if (
                      window.confirm(
                        `Delete ${ret.tax_year} tax return?`,
                      )
                    ) {
                      deleteReturn(ret.id)
                    }
                  }}
                  className="rounded-md p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                  aria-label="Delete return"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function formatFilingStatus(status: string): string {
  const labels: Record<string, string> = {
    single: "Single",
    mfj: "Married Filing Jointly",
    mfs: "Married Filing Separately",
    hoh: "Head of Household",
    qss: "Qualifying Surviving Spouse",
  }
  return labels[status] || status
}
