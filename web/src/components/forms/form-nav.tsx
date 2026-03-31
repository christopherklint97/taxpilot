import { useEffect } from "react"
import { useFormStore } from "@/stores/form-store"
import { useUIStore } from "@/stores/ui-store"
import { useFieldStore } from "@/stores/field-store"
import { cn } from "@/lib/utils"
import { FileText, Eye, EyeOff } from "lucide-react"

/** Display order for groups */
const GROUP_ORDER = [
  "personal",
  "income_w2",
  "income_1099",
  "business",
  "investments",
  "deductions",
  "credits",
  "taxes",
  "expat",
  "ca",
]

const GROUP_LABELS: Record<string, string> = {
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

/** Extract form number for display, e.g. "1040", "Schedule A", "8995" */
function formLabel(id: string, name: string): string {
  // For schedules, keep "Sch X" prefix
  const schedMatch = id.match(/^schedule_([a-z0-9]+)$/i)
  if (schedMatch) {
    const letter = schedMatch[1].toUpperCase()
    // Shorten the full name after the dash
    const dashIdx = name.indexOf("—")
    const shortName = dashIdx > 0 ? name.slice(dashIdx + 1).trim() : ""
    return shortName ? `Sch ${letter} — ${shortName}` : `Schedule ${letter}`
  }

  // CA forms
  if (id === "ca_540") return "540 — CA Resident Income Tax"
  if (id === "ca_schedule_ca") return "Sch CA — California Adjustments"

  // Numbered forms: form_XXXX → "XXXX — Name"
  const numMatch = id.match(/^(?:form_)?(\d+\w*)$/)
  if (numMatch) {
    const num = numMatch[1]
    const dashIdx = name.indexOf("—")
    const shortName = dashIdx > 0 ? name.slice(dashIdx + 1).trim() : name
    // Avoid repeating the number
    if (shortName.startsWith(num)) return shortName
    return `${num} — ${shortName}`
  }

  // Input forms (W-2, 1099-*)
  return name
}

export function FormNav() {
  const forms = useFormStore((s) => s.forms)
  const loadForms = useFormStore((s) => s.loadForms)
  const selectedFormId = useUIStore((s) => s.selectedFormId)
  const setSelectedForm = useUIStore((s) => s.setSelectedForm)
  const hiddenForms = useUIStore((s) => s.hiddenForms)
  const toggleFormHidden = useUIStore((s) => s.toggleFormHidden)
  const showEmptyForms = useUIStore((s) => s.showEmptyForms)
  const toggleShowEmptyForms = useUIStore((s) => s.toggleShowEmptyForms)
  const fields = useFieldStore((s) => s.fields)

  useEffect(() => {
    loadForms()
  }, [loadForms])

  // Determine which forms have data (any field with a nonzero/nonempty value)
  const formsWithData = new Set<string>()
  for (const [key, fv] of fields) {
    const formId = key.split(":")[0]
    if (
      (fv.value_num !== null && fv.value_num !== 0) ||
      (fv.value_str !== null && fv.value_str !== "")
    ) {
      formsWithData.add(formId)
    }
  }

  // Group forms in display order
  const groups = new Map<string, typeof forms>()
  for (const form of forms) {
    const group = form.question_group || "other"
    if (!groups.has(group)) groups.set(group, [])
    groups.get(group)!.push(form)
  }

  const sortedGroups = GROUP_ORDER.filter((g) => groups.has(g)).map(
    (g) => [g, groups.get(g)!] as const,
  )
  // Add any ungrouped
  for (const [g, gf] of groups) {
    if (!GROUP_ORDER.includes(g)) sortedGroups.push([g, gf])
  }

  const activeCount = forms.filter(
    (f) => formsWithData.has(f.id) && !hiddenForms.has(f.id),
  ).length

  return (
    <nav className="space-y-3 p-2">
      {/* Controls */}
      <div className="flex items-center justify-between px-2">
        <span className="text-[10px] font-medium text-muted-foreground">
          {activeCount} active forms
        </span>
        <button
          onClick={toggleShowEmptyForms}
          className="flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground"
          title={showEmptyForms ? "Hide empty forms" : "Show all forms"}
        >
          {showEmptyForms ? (
            <EyeOff className="h-3 w-3" />
          ) : (
            <Eye className="h-3 w-3" />
          )}
          {showEmptyForms ? "Hide empty" : "Show all"}
        </button>
      </div>

      {sortedGroups.map(([group, groupForms]) => {
        const visibleForms = groupForms.filter((form) => {
          if (hiddenForms.has(form.id)) return false
          if (!showEmptyForms && !formsWithData.has(form.id)) return false
          return true
        })
        if (visibleForms.length === 0) return null

        return (
          <div key={group}>
            <h4 className="mb-1 px-2 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
              {GROUP_LABELS[group] || group}
            </h4>
            <div className="space-y-0.5">
              {visibleForms.map((form) => {
                const hasData = formsWithData.has(form.id)
                return (
                  <div
                    key={form.id}
                    className="group flex items-center"
                  >
                    <button
                      onClick={() => setSelectedForm(form.id)}
                      className={cn(
                        "flex flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
                        selectedFormId === form.id
                          ? "bg-accent text-accent-foreground font-medium"
                          : hasData
                            ? "text-foreground hover:bg-accent/50"
                            : "text-muted-foreground/60 hover:bg-accent/30 hover:text-muted-foreground",
                      )}
                    >
                      <FileText
                        className={cn(
                          "h-3.5 w-3.5 shrink-0",
                          hasData ? "text-primary" : "opacity-40",
                        )}
                      />
                      <span className="truncate text-xs">
                        {formLabel(form.id, form.name)}
                      </span>
                      <span className="ml-auto text-[10px] opacity-50">
                        {form.field_count}
                      </span>
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        toggleFormHidden(form.id)
                      }}
                      className="hidden rounded p-0.5 text-muted-foreground/50 hover:text-muted-foreground group-hover:block"
                      title="Hide this form"
                    >
                      <EyeOff className="h-3 w-3" />
                    </button>
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </nav>
  )
}
