import { useEffect } from "react"
import { useFormStore } from "@/stores/form-store"
import { useUIStore } from "@/stores/ui-store"
import { cn } from "@/lib/utils"
import { FileText } from "lucide-react"

const FORM_GROUPS: Record<string, string> = {
  personal: "Personal Info",
  income_w2: "W-2 Income",
  income_1099: "1099 Income",
  business: "Business",
  investments: "Investments",
  deductions: "Deductions",
  credits: "Credits",
  taxes: "Taxes",
  expat: "Foreign Income",
  ca: "California",
}

export function FormNav() {
  const forms = useFormStore((s) => s.forms)
  const loadForms = useFormStore((s) => s.loadForms)
  const selectedFormId = useUIStore((s) => s.selectedFormId)
  const setSelectedForm = useUIStore((s) => s.setSelectedForm)

  useEffect(() => {
    loadForms()
  }, [loadForms])

  // Group forms
  const groups = new Map<string, typeof forms>()
  for (const form of forms) {
    const group = form.question_group || "other"
    if (!groups.has(group)) groups.set(group, [])
    groups.get(group)!.push(form)
  }

  return (
    <nav className="space-y-4 p-2">
      {Array.from(groups.entries()).map(([group, groupForms]) => (
        <div key={group}>
          <h4 className="mb-1 px-2 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            {FORM_GROUPS[group] || group}
          </h4>
          <div className="space-y-0.5">
            {groupForms.map((form) => (
              <button
                key={form.id}
                onClick={() => setSelectedForm(form.id)}
                className={cn(
                  "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
                  selectedFormId === form.id
                    ? "bg-accent text-accent-foreground font-medium"
                    : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                )}
              >
                <FileText className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{form.name}</span>
                <span className="ml-auto text-[10px] opacity-60">{form.field_count}</span>
              </button>
            ))}
          </div>
        </div>
      ))}
    </nav>
  )
}
