import { useState, useCallback, forwardRef } from "react"
import { cn } from "@/lib/utils"
import type { FieldMeta } from "@/api/types"
import { useFieldStore } from "@/stores/field-store"
import { PriorYearBadge } from "./prior-year-badge"
import { FieldHelp } from "./field-help"

interface FieldEditorProps {
  field: FieldMeta
  className?: string
  onTab?: () => void
}

export const FieldEditor = forwardRef<HTMLDivElement, FieldEditorProps>(
  function FieldEditor({ field, className, onTab }, ref) {
    const getNumeric = useFieldStore((s) => s.getNumeric)
    const getString = useFieldStore((s) => s.getString)
    const updateField = useFieldStore((s) => s.updateField)
    const pendingKeys = useFieldStore((s) => s.pendingKeys)

    const isPending = pendingKeys.has(field.field_key)
    const isComputed = field.field_type !== "user_input"
    const isString = field.value_type === "string"

    const numValue = getNumeric(field.field_key)
    const strValue = getString(field.field_key)

    const [localValue, setLocalValue] = useState<string>("")
    const [focused, setFocused] = useState(false)

    const displayValue = focused
      ? localValue
      : isString
        ? strValue
        : numValue !== 0
          ? formatDisplay(numValue, field.value_type)
          : ""

    const handleFocus = useCallback(() => {
      setFocused(true)
      setLocalValue(
        isString ? strValue : numValue !== 0 ? String(numValue) : "",
      )
    }, [isString, strValue, numValue])

    const handleBlur = useCallback(() => {
      setFocused(false)
      if (isString) {
        if (localValue !== strValue) {
          updateField(field.field_key, 0, localValue)
        }
      } else {
        const parsed = parseFloat(localValue) || 0
        if (parsed !== numValue) {
          updateField(field.field_key, parsed, null)
        }
      }
    }, [
      localValue,
      isString,
      strValue,
      numValue,
      field.field_key,
      updateField,
    ])

    const handleKeyDown = useCallback(
      (e: React.KeyboardEvent) => {
        if (e.key === "Enter") {
          ;(e.target as HTMLInputElement).blur()
        }
        if (e.key === "Tab" && !e.shiftKey && onTab) {
          e.preventDefault()
          ;(e.target as HTMLInputElement).blur()
          // Small delay for blur to process
          setTimeout(() => onTab(), 50)
        }
      },
      [onTab],
    )

    // Enum field (dropdown)
    if (field.options.length > 0) {
      return (
        <div ref={ref} className={cn("space-y-1", className)}>
          <label className="flex items-center gap-1 text-xs font-medium text-foreground">
            {field.label}
            <FieldHelp fieldKey={field.field_key} />
          </label>
          <select
            value={strValue}
            onChange={(e) => {
              updateField(field.field_key, 0, e.target.value)
            }}
            className="w-full rounded-md border-2 border-primary/30 bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
          >
            <option value="">Select...</option>
            {field.options.map((opt) => (
              <option key={opt} value={opt}>
                {formatOption(opt)}
              </option>
            ))}
          </select>
        </div>
      )
    }

    // Computed field (read-only display)
    if (isComputed) {
      const typeLabel =
        field.field_type === "federal_ref" ? "FROM FEDERAL" : "COMPUTED"
      return (
        <div ref={ref} className={cn("space-y-1", className)}>
          <label className="flex items-center gap-1 text-xs text-muted-foreground">
            {field.label}
            <FieldHelp fieldKey={field.field_key} />
            <span className="ml-1 rounded bg-muted px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider">
              {typeLabel}
            </span>
            <PriorYearBadge fieldKey={field.field_key} />
          </label>
          <div
            className={cn(
              "w-full rounded-md border border-transparent bg-muted/40 px-3 py-2 text-sm tabular-nums text-muted-foreground",
              isPending && "animate-pulse",
              numValue < 0 && "text-destructive",
            )}
          >
            {isString
              ? strValue || "—"
              : numValue !== 0
                ? formatDisplay(numValue, field.value_type)
                : "—"}
          </div>
        </div>
      )
    }

    // User input field (editable) — visually distinct with border + label color
    return (
      <div ref={ref} className={cn("space-y-1", className)}>
        <label className="flex items-center gap-1 text-xs font-medium text-foreground">
          {field.label}
          <FieldHelp fieldKey={field.field_key} />
        </label>
        <input
          type="text"
          inputMode={isString ? "text" : "decimal"}
          value={displayValue}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={(e) => setLocalValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={field.prompt || field.label}
          className={cn(
            "w-full rounded-md border-2 border-primary/30 bg-background px-3 py-2 text-sm tabular-nums",
            "placeholder:text-muted-foreground/40",
            "focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary",
            isPending && "border-primary/60 bg-primary/5",
          )}
        />
      </div>
    )
  },
)

function formatDisplay(value: number, valueType: string): string {
  if (value === 0) return ""
  if (valueType === "integer") return Math.round(value).toString()
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value)
}

function formatOption(opt: string): string {
  const labels: Record<string, string> = {
    single: "Single",
    mfj: "Married Filing Jointly",
    mfs: "Married Filing Separately",
    hoh: "Head of Household",
    qss: "Qualifying Surviving Spouse",
    ppt: "Physical Presence Test",
    bfrt: "Bona Fide Residence Test",
    general: "General Category",
    passive: "Passive Category",
    yes: "Yes",
    no: "No",
    short: "Short-term",
    long: "Long-term",
    accrued: "Accrued",
    paid: "Paid",
    deposit: "Deposit Account",
    custodial: "Custodial Account",
    other: "Other",
  }
  return labels[opt] || opt.charAt(0).toUpperCase() + opt.slice(1)
}
