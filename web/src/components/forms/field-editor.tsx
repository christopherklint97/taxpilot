import { useState, useCallback } from "react"
import { cn } from "@/lib/utils"
import type { FieldMeta } from "@/api/types"
import { useFieldStore } from "@/stores/field-store"
import { PriorYearBadge } from "./prior-year-badge"
import { FieldHelp } from "./field-help"

interface FieldEditorProps {
  field: FieldMeta
  className?: string
}

export function FieldEditor({ field, className }: FieldEditorProps) {
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
      isString ? strValue : numValue !== 0 ? String(numValue) : ""
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
  }, [localValue, isString, strValue, numValue, field.field_key, updateField])

  // Enum field (dropdown)
  if (field.options.length > 0) {
    return (
      <div className={cn("space-y-1", className)}>
        <label className="flex items-center gap-1 text-xs font-medium text-muted-foreground">
          {field.label}
          <FieldHelp fieldKey={field.field_key} />
        </label>
        <select
          value={strValue}
          onChange={(e) => {
            updateField(field.field_key, 0, e.target.value)
          }}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
        >
          <option value="">Select...</option>
          {field.options.map((opt) => (
            <option key={opt} value={opt}>{formatOption(opt)}</option>
          ))}
        </select>
      </div>
    )
  }

  // Computed field (read-only display)
  if (isComputed) {
    return (
      <div className={cn("space-y-1", className)}>
        <label className="flex items-center gap-1 text-xs font-medium text-muted-foreground">
          {field.label}
          <FieldHelp fieldKey={field.field_key} />
          <span className="ml-1 text-[10px] uppercase tracking-wide opacity-60">
            {field.field_type === "federal_ref" ? "from federal" : "computed"}
          </span>
          <PriorYearBadge fieldKey={field.field_key} />
        </label>
        <div className={cn(
          "w-full rounded-md border border-input/50 bg-muted/30 px-3 py-2 text-sm tabular-nums",
          isPending && "animate-pulse",
          numValue < 0 && "text-destructive"
        )}>
          {isString ? strValue : formatDisplay(numValue, field.value_type)}
        </div>
      </div>
    )
  }

  // User input field (editable)
  return (
    <div className={cn("space-y-1", className)}>
      <label className="flex items-center gap-1 text-xs font-medium text-muted-foreground">
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
        onKeyDown={(e) => {
          if (e.key === "Enter") (e.target as HTMLInputElement).blur()
        }}
        placeholder={field.prompt || field.label}
        className={cn(
          "w-full rounded-md border border-input bg-background px-3 py-2 text-sm tabular-nums",
          "focus:outline-none focus:ring-2 focus:ring-ring",
          isPending && "border-primary/50",
        )}
      />
    </div>
  )
}

function formatDisplay(value: number, valueType: string): string {
  if (value === 0) return ""
  if (valueType === "integer") return Math.round(value).toString()
  // Currency format with commas, no cents
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
