import { useState, useEffect, useRef, useCallback } from "react"
import { PdfViewer } from "./pdf-viewer"
import { FormSection } from "@/components/forms/form-section"
import { useFieldStore } from "@/stores/field-store"
import { cn } from "@/lib/utils"
import { PanelLeftClose, PanelLeftOpen } from "lucide-react"

interface PdfSideBySideProps {
  formId: string
}

export function PdfSideBySide({ formId }: PdfSideBySideProps) {
  const returnId = useFieldStore((s) => s.returnId)
  const solveVersion = useFieldStore((s) => s.solveVersion)

  const [pdfData, setPdfData] = useState<Uint8Array | null>(null)
  const [pdfLoading, setPdfLoading] = useState(false)
  const [editorCollapsed, setEditorCollapsed] = useState(false)

  // Debounce timer ref
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  // Track the version we last fetched for
  const lastFetchedVersion = useRef(-1)

  const fetchPdf = useCallback(async () => {
    if (!returnId) return
    setPdfLoading(true)
    try {
      const res = await fetch(
        `/api/returns/${returnId}/pdf/filled/${formId}`,
      )
      if (!res.ok) {
        setPdfData(null)
        return
      }
      const contentType = res.headers.get("content-type") || ""
      if (contentType.includes("application/pdf")) {
        const buf = await res.arrayBuffer()
        setPdfData(new Uint8Array(buf))
      } else {
        // Server returned JSON fallback (no template)
        setPdfData(null)
      }
    } catch {
      setPdfData(null)
    } finally {
      setPdfLoading(false)
    }
  }, [returnId, formId])

  // Fetch PDF on mount
  useEffect(() => {
    fetchPdf()
    lastFetchedVersion.current = solveVersion
  }, [formId, returnId, fetchPdf])

  // Debounced refresh when solver produces new results
  useEffect(() => {
    if (solveVersion === lastFetchedVersion.current) return

    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      lastFetchedVersion.current = solveVersion
      fetchPdf()
    }, 500)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [solveVersion, fetchPdf])

  return (
    <div className="flex h-full">
      {/* PDF panel */}
      <div
        className={cn(
          "flex flex-col border-r border-border transition-all",
          editorCollapsed ? "flex-1" : "w-1/2",
        )}
      >
        <div className="flex items-center justify-between border-b border-border bg-muted/30 px-3 py-1.5">
          <span className="text-xs font-medium text-muted-foreground">
            PDF Preview — {formId.toUpperCase()}
          </span>
          <button
            onClick={() => setEditorCollapsed((c) => !c)}
            className="rounded p-1 hover:bg-accent"
            aria-label={
              editorCollapsed ? "Show editor" : "Expand PDF"
            }
          >
            {editorCollapsed ? (
              <PanelLeftOpen className="h-4 w-4" />
            ) : (
              <PanelLeftClose className="h-4 w-4" />
            )}
          </button>
        </div>
        <PdfViewer
          data={pdfData}
          loading={pdfLoading}
          className="flex-1"
        />
      </div>

      {/* Editor panel */}
      {!editorCollapsed && (
        <div className="w-1/2 overflow-y-auto p-6">
          <FormSection formId={formId} />
        </div>
      )}
    </div>
  )
}
