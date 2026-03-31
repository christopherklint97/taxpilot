import { useEffect } from "react"
import { useFormStore } from "@/stores/form-store"
import { FieldEditor } from "./field-editor"
import { Loader2 } from "lucide-react"

interface FormSectionProps {
  formId: string
}

export function FormSection({ formId }: FormSectionProps) {
  const loadFormDetail = useFormStore((s) => s.loadFormDetail)
  const formDetail = useFormStore((s) => s.getFormDetail(formId))

  useEffect(() => {
    loadFormDetail(formId)
  }, [formId, loadFormDetail])

  if (!formDetail) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  // Group fields: user inputs first, then computed
  const userInputs = formDetail.fields.filter(f => f.field_type === "user_input")
  const computed = formDetail.fields.filter(f => f.field_type !== "user_input")

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold">{formDetail.name}</h3>
        <p className="text-sm text-muted-foreground">
          {formDetail.jurisdiction === "federal" ? "Federal" : "California"} &middot;{" "}
          {userInputs.length} inputs, {computed.length} computed
        </p>
      </div>

      {userInputs.length > 0 && (
        <div className="space-y-1">
          <h4 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Your Input
          </h4>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {userInputs.map((field) => (
              <FieldEditor key={field.field_key} field={field} />
            ))}
          </div>
        </div>
      )}

      {computed.length > 0 && (
        <div className="space-y-1">
          <h4 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Computed Values
          </h4>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {computed.map((field) => (
              <FieldEditor key={field.field_key} field={field} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
