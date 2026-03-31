import { create } from "zustand"
import type { FieldValue, ChangedField, TaxReturnDetail, PriorYearValue } from "@/api/types"
import { updateField as apiUpdateField } from "@/api/hooks"

interface FieldStore {
  returnId: string | null
  taxYear: number
  fields: Map<string, FieldValue>
  pendingKeys: Set<string>
  /** Incremented on every solver response — used to trigger PDF refresh */
  solveVersion: number
  /** Prior year numeric values for delta comparison */
  priorValues: Map<string, number>

  // Load fields from a return detail response
  loadReturn: (detail: TaxReturnDetail) => void

  // Get a field value
  getField: (key: string) => FieldValue | undefined
  getNumeric: (key: string) => number
  getString: (key: string) => string

  // Prior year values
  loadPriorValues: (values: PriorYearValue[]) => void
  getPriorValue: (key: string) => number | undefined

  // Update a field and trigger solver
  updateField: (key: string, valueNum: number | null, valueStr: string | null) => Promise<void>

  // Apply solver response (changed computed fields)
  applyChanges: (changes: ChangedField[]) => void

  // Clear store
  clear: () => void
}

export const useFieldStore = create<FieldStore>((set, get) => ({
  returnId: null,
  taxYear: 2025,
  fields: new Map(),
  pendingKeys: new Set(),
  solveVersion: 0,
  priorValues: new Map(),

  loadReturn: (detail) => {
    const fields = new Map<string, FieldValue>()
    for (const f of detail.fields) {
      fields.set(f.field_key, f)
    }
    set({
      returnId: detail.id,
      taxYear: detail.tax_year,
      fields,
    })
  },

  loadPriorValues: (values) => {
    const priorValues = new Map<string, number>()
    for (const v of values) {
      if (v.value_num !== null) {
        priorValues.set(v.field_key, v.value_num)
      }
    }
    set({ priorValues })
  },

  getPriorValue: (key) => get().priorValues.get(key),

  getField: (key) => get().fields.get(key),

  getNumeric: (key) => {
    const f = get().fields.get(key)
    return f?.value_num ?? 0
  },

  getString: (key) => {
    const f = get().fields.get(key)
    return f?.value_str ?? ""
  },

  updateField: async (key, valueNum, valueStr) => {
    const { returnId, fields, pendingKeys } = get()
    if (!returnId) return

    // Optimistic update
    const current = fields.get(key)
    const updated: FieldValue = {
      field_key: key,
      value_num: valueNum ?? current?.value_num ?? null,
      value_str: valueStr ?? current?.value_str ?? null,
      source: "user_input",
    }
    const newFields = new Map(fields)
    newFields.set(key, updated)
    const newPending = new Set(pendingKeys)
    newPending.add(key)
    set({ fields: newFields, pendingKeys: newPending })

    try {
      const response = await apiUpdateField(returnId, key, valueNum, valueStr)
      // Apply computed field changes from solver
      get().applyChanges(response.changed_fields)
    } catch {
      // Revert on error
      if (current) {
        const revertFields = new Map(get().fields)
        revertFields.set(key, current)
        set({ fields: revertFields })
      }
    } finally {
      const finalPending = new Set(get().pendingKeys)
      finalPending.delete(key)
      set({ pendingKeys: finalPending })
    }
  },

  applyChanges: (changes) => {
    const newFields = new Map(get().fields)
    for (const change of changes) {
      const existing = newFields.get(change.key)
      newFields.set(change.key, {
        field_key: change.key,
        value_num: change.value_num,
        value_str: change.value_str,
        source: existing?.source ?? "computed",
      })
    }
    set({ fields: newFields, solveVersion: get().solveVersion + 1 })
  },

  clear: () => {
    set({ returnId: null, fields: new Map(), pendingKeys: new Set(), priorValues: new Map() })
  },
}))
