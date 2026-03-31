import { useEffect, useRef, useCallback } from "react"
import { useFormStore } from "@/stores/form-store"
import { useUIStore } from "@/stores/ui-store"
import { FieldEditor } from "./field-editor"
import { Loader2, Filter, ChevronDown } from "lucide-react"

interface FormSectionProps {
  formId: string
}

export function FormSection({ formId }: FormSectionProps) {
  const loadFormDetail = useFormStore((s) => s.loadFormDetail)
  const formDetail = useFormStore((s) => s.getFormDetail(formId))
  const inputsOnly = useUIStore((s) => s.inputsOnly)
  const toggleInputsOnly = useUIStore((s) => s.toggleInputsOnly)

  const inputRefs = useRef<Map<string, HTMLElement>>(new Map())

  useEffect(() => {
    loadFormDetail(formId)
  }, [formId, loadFormDetail])

  const registerRef = useCallback(
    (key: string, el: HTMLElement | null) => {
      if (el) inputRefs.current.set(key, el)
      else inputRefs.current.delete(key)
    },
    [],
  )

  // Navigate to next input field
  const focusNextInput = useCallback(
    (currentKey: string) => {
      if (!formDetail) return
      const inputs = formDetail.fields.filter(
        (f) => f.field_type === "user_input",
      )
      const idx = inputs.findIndex((f) => f.field_key === currentKey)
      if (idx >= 0 && idx < inputs.length - 1) {
        const nextKey = inputs[idx + 1].field_key
        const el = inputRefs.current.get(nextKey)
        if (el) {
          el.scrollIntoView({ behavior: "smooth", block: "center" })
          const input = el.querySelector("input, select") as HTMLElement
          input?.focus()
        }
      }
    },
    [formDetail],
  )

  if (!formDetail) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const userInputs = formDetail.fields.filter(
    (f) => f.field_type === "user_input",
  )
  const computed = formDetail.fields.filter(
    (f) => f.field_type !== "user_input",
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold">{formDetail.name}</h3>
          <p className="text-sm text-muted-foreground">
            {formDetail.jurisdiction === "federal" ? "Federal" : "California"}{" "}
            &middot; {userInputs.length} inputs, {computed.length} computed
          </p>
        </div>
        <button
          onClick={toggleInputsOnly}
          className={`flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors ${
            inputsOnly
              ? "border-primary bg-primary/10 text-primary"
              : "border-border text-muted-foreground hover:bg-accent"
          }`}
        >
          <Filter className="h-3 w-3" />
          {inputsOnly ? "Inputs only" : "All fields"}
        </button>
      </div>

      {/* User inputs */}
      {userInputs.length > 0 && (
        <div className="space-y-1">
          <h4 className="text-sm font-medium uppercase tracking-wide text-primary">
            Your Input
          </h4>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {userInputs.map((field) => (
              <FieldEditor
                key={field.field_key}
                field={field}
                ref={(el) => registerRef(field.field_key, el)}
                onTab={() => focusNextInput(field.field_key)}
              />
            ))}
          </div>
        </div>
      )}

      {/* Computed values */}
      {!inputsOnly && computed.length > 0 && (
        <div className="space-y-1">
          <h4 className="flex items-center gap-1 text-sm font-medium uppercase tracking-wide text-muted-foreground">
            Computed Values
            <ChevronDown className="h-3 w-3" />
          </h4>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {computed.map((field) => (
              <FieldEditor key={field.field_key} field={field} />
            ))}
          </div>
        </div>
      )}

      {inputsOnly && computed.length > 0 && (
        <p className="text-xs text-muted-foreground">
          {computed.length} computed fields hidden.{" "}
          <button
            onClick={toggleInputsOnly}
            className="underline hover:text-foreground"
          >
            Show all
          </button>
        </p>
      )}
    </div>
  )
}
