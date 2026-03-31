import { useState, useEffect, useCallback, useRef, useMemo } from "react"
import { Document, Page, pdfjs } from "react-pdf"
import "react-pdf/dist/Page/AnnotationLayer.css"
import "react-pdf/dist/Page/TextLayer.css"
import { ChevronLeft, ChevronRight, ZoomIn, ZoomOut, Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

// Configure pdf.js worker
pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`

interface PdfViewerProps {
  /** PDF data — either a URL string or Uint8Array blob */
  data: string | Uint8Array | null
  /** Loading state */
  loading?: boolean
  className?: string
}

export function PdfViewer({ data, loading, className }: PdfViewerProps) {
  const [numPages, setNumPages] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [scale, setScale] = useState(1.0)
  const [pdfError, setPdfError] = useState<string | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const onDocumentLoadSuccess = useCallback(
    ({ numPages: n }: { numPages: number }) => {
      setNumPages(n)
      setCurrentPage(1)
      setPdfError(null)
    },
    [],
  )

  const onDocumentLoadError = useCallback((error: Error) => {
    setPdfError(error.message)
  }, [])

  // Copy the buffer so pdf.js can transfer it to the worker without detaching
  // our source array. Memoize to avoid reloading on unrelated re-renders.
  const file = useMemo(
    () => (data instanceof Uint8Array ? { data: new Uint8Array(data) } : data),
    [data],
  )

  // Reset page when data changes
  useEffect(() => {
    setCurrentPage(1)
    setPdfError(null)
  }, [data])

  if (loading) {
    return (
      <div
        className={cn(
          "flex items-center justify-center bg-muted/20",
          className,
        )}
      >
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <Loader2 className="h-8 w-8 animate-spin" />
          <span className="text-sm">Generating PDF...</span>
        </div>
      </div>
    )
  }

  if (!data) {
    return (
      <div
        className={cn(
          "flex items-center justify-center bg-muted/20",
          className,
        )}
      >
        <div className="text-center text-muted-foreground">
          <p className="text-sm">No PDF template available.</p>
          <p className="mt-1 text-xs">
            Add PDF templates to data/tax_years/ to enable preview.
          </p>
        </div>
      </div>
    )
  }

  if (pdfError) {
    return (
      <div
        className={cn(
          "flex items-center justify-center bg-muted/20",
          className,
        )}
      >
        <div className="text-center text-muted-foreground">
          <p className="text-sm">Failed to render PDF</p>
          <p className="mt-1 text-xs text-destructive">{pdfError}</p>
        </div>
      </div>
    )
  }

  return (
    <div className={cn("flex flex-col", className)}>
      {/* Toolbar */}
      <div className="flex items-center justify-between border-b border-border bg-muted/30 px-3 py-1.5">
        <div className="flex items-center gap-1">
          <button
            onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
            disabled={currentPage <= 1}
            className="rounded p-1 hover:bg-accent disabled:opacity-30"
            aria-label="Previous page"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <span className="min-w-[4rem] text-center text-xs text-muted-foreground">
            {currentPage} / {numPages || "–"}
          </span>
          <button
            onClick={() => setCurrentPage((p) => Math.min(numPages, p + 1))}
            disabled={currentPage >= numPages}
            className="rounded p-1 hover:bg-accent disabled:opacity-30"
            aria-label="Next page"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={() => setScale((s) => Math.max(0.5, s - 0.15))}
            className="rounded p-1 hover:bg-accent"
            aria-label="Zoom out"
          >
            <ZoomOut className="h-4 w-4" />
          </button>
          <span className="min-w-[3rem] text-center text-xs text-muted-foreground">
            {Math.round(scale * 100)}%
          </span>
          <button
            onClick={() => setScale((s) => Math.min(2.5, s + 0.15))}
            className="rounded p-1 hover:bg-accent"
            aria-label="Zoom in"
          >
            <ZoomIn className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* PDF content */}
      <div ref={containerRef} className="flex-1 overflow-auto bg-neutral-200 dark:bg-neutral-800 p-4">
        <Document
          file={file}
          onLoadSuccess={onDocumentLoadSuccess}
          onLoadError={onDocumentLoadError}
          loading={
            <div className="flex justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          }
        >
          <Page
            pageNumber={currentPage}
            scale={scale}
            renderTextLayer={true}
            renderAnnotationLayer={true}
          />
        </Document>
      </div>
    </div>
  )
}
