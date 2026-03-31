import { useEffect, useRef, useState } from "react"
import { useNavigate } from "@tanstack/react-router"
import { useReturnStore } from "@/stores/return-store"
import { rollforward } from "@/api/hooks"
import { FileText, Plus, Trash2, ArrowRight, Upload } from "lucide-react"

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
      // Create a new 2024 return first, then upload PDF to it
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

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Tax Returns</h2>
          <p className="text-sm text-muted-foreground">
            Manage your federal and state tax returns
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
            {creating ? "Creating..." : "New Return"}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
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
          <p className="mb-4 text-sm text-muted-foreground">
            Create a new 2025 return or import your 2024 PDF to get started.
          </p>
          <div className="flex items-center justify-center gap-3">
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
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" />
              {creating ? "Creating..." : "Create 2025 Return"}
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {returns.map((ret) => (
            <div
              key={ret.id}
              className="flex items-center justify-between rounded-lg border border-border p-4 hover:bg-accent/50 cursor-pointer"
              onClick={() =>
                navigate({
                  to: "/returns/$returnId",
                  params: { returnId: ret.id },
                })
              }
            >
              <div className="flex items-center gap-4">
                <FileText className="h-8 w-8 text-muted-foreground" />
                <div>
                  <p className="font-medium">
                    {ret.tax_year} Tax Return — {ret.state_code}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {ret.filing_status ?? "Not started"} &middot; Updated{" "}
                    {new Date(ret.updated_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={async (e) => {
                    e.stopPropagation()
                    setRollingForward(ret.id)
                    try {
                      const result = await rollforward(ret.id)
                      await fetchReturns()
                      navigate({
                        to: "/returns/$returnId",
                        params: { returnId: result.return_id },
                      })
                    } finally {
                      setRollingForward(null)
                    }
                  }}
                  disabled={rollingForward === ret.id}
                  className="rounded-md p-2 text-muted-foreground hover:bg-primary/10 hover:text-primary disabled:opacity-50"
                  aria-label="Roll forward to next year"
                  title={`Roll forward to ${ret.tax_year + 1}`}
                >
                  <ArrowRight className="h-4 w-4" />
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    deleteReturn(ret.id)
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
