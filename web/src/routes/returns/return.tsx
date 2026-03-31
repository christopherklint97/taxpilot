import { useEffect, useState } from "react"
import { useParams, useNavigate } from "@tanstack/react-router"
import { useFieldStore } from "@/stores/field-store"
import { useUIStore } from "@/stores/ui-store"
import { useFormStore } from "@/stores/form-store"
import { fetchReturn, fetchPriorYearValues } from "@/api/hooks"
import { FormNav } from "@/components/forms/form-nav"
import { FormSection } from "@/components/forms/form-section"
import { PdfSideBySide } from "@/components/pdf/pdf-side-by-side"
import { ValidationPanel as ValidationPanelLazy } from "@/components/validation/validation-panel"
import { ArrowLeft, Loader2 } from "lucide-react"

type ViewMode = "editor" | "pdf" | "validate"

export function ReturnPage() {
  const { returnId } = useParams({ from: "/returns/$returnId" })
  const navigate = useNavigate()

  const loadReturn = useFieldStore((s) => s.loadReturn)
  const loadPriorValues = useFieldStore((s) => s.loadPriorValues)
  const clearFields = useFieldStore((s) => s.clear)
  const taxYear = useFieldStore((s) => s.taxYear)

  const selectedFormId = useUIStore((s) => s.selectedFormId)
  const setSelectedForm = useUIStore((s) => s.setSelectedForm)

  const loadForms = useFormStore((s) => s.loadForms)

  const [viewMode, setViewMode] = useState<ViewMode>("editor")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Load return data
  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    Promise.all([
      fetchReturn(returnId),
      loadForms(),
    ])
      .then(async ([detail]) => {
        if (cancelled) return
        loadReturn(detail)
        // Select first form if none selected
        if (!selectedFormId) setSelectedForm("1040")
        // Load prior year values (non-blocking)
        try {
          const pv = await fetchPriorYearValues(returnId)
          if (!cancelled) loadPriorValues(pv)
        } catch {
          // Prior year values optional
        }
      })
      .catch((e) => {
        if (!cancelled) setError((e as Error).message)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
      clearFields()
    }
  }, [returnId])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <p className="text-destructive">{error}</p>
          <button
            onClick={() => navigate({ to: "/" })}
            className="mt-4 text-sm text-muted-foreground hover:text-foreground"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full">
      {/* Form navigation sidebar */}
      <div className="w-64 shrink-0 overflow-y-auto border-r border-border bg-background">
        <div className="sticky top-0 z-10 border-b border-border bg-background p-3">
          <button
            onClick={() => navigate({ to: "/" })}
            className="mb-2 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-3 w-3" /> Dashboard
          </button>
          <h2 className="text-sm font-semibold">{taxYear} Tax Return</h2>
          <div className="mt-2 flex gap-1">
            {(["editor", "pdf", "validate"] as ViewMode[]).map((mode) => (
              <button
                key={mode}
                onClick={() => setViewMode(mode)}
                className={`rounded-md px-2.5 py-1 text-xs font-medium capitalize ${
                  viewMode === mode
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground hover:bg-accent/50"
                }`}
              >
                {mode === "pdf" ? "PDF" : mode === "validate" ? "Validate" : "Editor"}
              </button>
            ))}
          </div>
        </div>
        <FormNav />
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        {viewMode === "editor" && selectedFormId && (
          <div className="h-full overflow-y-auto p-6">
            <FormSection formId={selectedFormId} />
          </div>
        )}
        {viewMode === "pdf" && selectedFormId && (
          <PdfSideBySide formId={selectedFormId} />
        )}
        {viewMode === "validate" && (
          <div className="h-full overflow-y-auto p-6">
            <ValidationPanelLazy />
          </div>
        )}
        {!selectedFormId && (
          <div className="flex h-full items-center justify-center text-muted-foreground">
            Select a form from the sidebar to begin
          </div>
        )}
      </div>
    </div>
  )
}
