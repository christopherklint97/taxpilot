import { useState, useCallback } from "react"
import { explainField } from "@/api/hooks"
import type { ExplainResponse } from "@/api/types"
import { HelpCircle, Loader2, X } from "lucide-react"
import { cn } from "@/lib/utils"

interface FieldHelpProps {
  fieldKey: string
  className?: string
}

export function FieldHelp({ fieldKey, className }: FieldHelpProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ExplainResponse | null>(null)

  const handleClick = useCallback(async () => {
    if (open) {
      setOpen(false)
      return
    }
    setOpen(true)
    if (result) return // already loaded

    setLoading(true)
    try {
      const res = await explainField(fieldKey)
      setResult(res)
    } catch {
      setResult({
        explanation: "Failed to load explanation.",
        model: "",
        configured: false,
      })
    } finally {
      setLoading(false)
    }
  }, [open, result, fieldKey])

  return (
    <div className={cn("relative inline-flex", className)}>
      <button
        onClick={handleClick}
        className="rounded p-0.5 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
        aria-label="Explain this field"
        title="Get AI explanation"
      >
        <HelpCircle className="h-3.5 w-3.5" />
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-1 w-72 rounded-lg border border-border bg-popover p-3 shadow-lg">
          <div className="flex items-start justify-between gap-2">
            <span className="text-[11px] font-medium text-muted-foreground">
              AI Explanation
            </span>
            <button
              onClick={() => setOpen(false)}
              className="rounded p-0.5 hover:bg-accent"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
          {loading ? (
            <div className="flex items-center gap-2 py-3 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Generating explanation...
            </div>
          ) : result ? (
            <div className="mt-2 space-y-2">
              <p className="text-sm leading-relaxed text-foreground whitespace-pre-wrap">
                {result.explanation}
              </p>
              {result.model && (
                <p className="text-[10px] text-muted-foreground">
                  via {result.model}
                </p>
              )}
            </div>
          ) : null}
        </div>
      )}
    </div>
  )
}
