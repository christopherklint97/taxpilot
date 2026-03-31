import { useFieldStore } from "@/stores/field-store"
import { cn } from "@/lib/utils"
import { TrendingUp, TrendingDown, Minus } from "lucide-react"

interface PriorYearBadgeProps {
  fieldKey: string
}

export function PriorYearBadge({ fieldKey }: PriorYearBadgeProps) {
  const currentValue = useFieldStore((s) => s.getNumeric(fieldKey))
  const priorValue = useFieldStore((s) => s.getPriorValue(fieldKey))

  if (priorValue === undefined) return null

  const delta = currentValue - priorValue
  if (Math.abs(delta) < 1) return null

  const pct = priorValue !== 0 ? (delta / priorValue) * 100 : 0
  const isUp = delta > 0
  const isDown = delta < 0

  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[10px] font-medium",
        isUp &&
          "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
        isDown &&
          "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
        !isUp &&
          !isDown &&
          "bg-muted text-muted-foreground",
      )}
    >
      {isUp && <TrendingUp className="h-2.5 w-2.5" />}
      {isDown && <TrendingDown className="h-2.5 w-2.5" />}
      {!isUp && !isDown && <Minus className="h-2.5 w-2.5" />}
      {isUp ? "+" : ""}
      {formatDelta(delta)}
      {priorValue !== 0 && (
        <span className="opacity-60">
          ({pct > 0 ? "+" : ""}
          {pct.toFixed(1)}%)
        </span>
      )}
    </span>
  )
}

function formatDelta(value: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value)
}
